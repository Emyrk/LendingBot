package scraper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/DistributedSolutions/LendingBot/app/jsonrpc"
	"github.com/fatih/color"
)

var _ = ioutil.ReadDir

const (
	invalidParameters uint8 = iota
	customError       uint8 = iota
	noError           uint8 = iota
)

func (p *Scraper) Serve() {
	closer := ServeRouter(p.Router, 8080)
	p.apicloser = closer
}

func (p *Scraper) Close() {
	p.apicloser.Close()
}

type ApiBase struct{}

//method that all of api methods will be tested against
func (a ApiBase) ApiBaseMethod(json json.RawMessage) (successResponse *interface{}, apiError *ApiError, errorType uint8) {
	return nil, nil, noError
}

//used to provide the methods for api calls
type ApiProvider struct {
	Scraper *Scraper
}

type ApiService struct {
	Scraper *Scraper
	Api     ApiProvider
}

func Vars(r *http.Request) map[string]string {
	return make(map[string]string)
}

func marshalErr(err *jsonrpc.JSONRPCReponse) []byte {
	data, _ := err.CustomMarshalJSON()
	return data
}

func (apiService *ApiService) HandleAPICalls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(marshalErr(jsonrpc.NewInternalRPCSError("Error reading the body of the request", 0)))
		return
	}

	var extra string
	var errorID uint32

	req := jsonrpc.NewEmptyJSONRPCRequest()
	err = json.Unmarshal(data, req)
	if err != nil {
		w.Write(marshalErr(jsonrpc.NewParseError(err.Error(), 0)))
		return
	}

	resp := new(jsonrpc.JSONRPCReponse)
	resp.Id = req.ID
	var result json.RawMessage

	methodCamelCase := DashDelimiterToCamelCase(req.Method)
	color.Green(fmt.Sprintf("%s: %s: %s", time.Now().Format("15:04:05"), r.Method, methodCamelCase))

	apiProvider := ApiProvider{
		apiService.Scraper,
	}

	//get the api interface object
	_, ok := reflect.TypeOf(apiProvider).MethodByName(methodCamelCase)
	//checks to see if method exists to call it
	if !ok {
		//method does not exist
		extra = fmt.Sprintf("Method not found: %s", methodCamelCase)
		color.Red(extra)
		goto MethodNotFound
	} else {
		method := reflect.ValueOf(apiProvider).MethodByName(methodCamelCase)
		//method exists
		//successResponse JSON before
		//apiError = object containing messages
		//errorType:
		//		InvalidParameters
		//		CustomError
		in := []reflect.Value{reflect.ValueOf(req.Params)}
		resultValues := method.Call(in)

		if !resultValues[1].CanInterface() || !resultValues[2].CanInterface() {
			extra = fmt.Sprintf("Unable to interface either 1: %d, or 2: %d of the return values.", resultValues[1].CanInterface(), resultValues[2].CanInterface())
			color.Red(extra)
			goto InternalError
		}
		//if the apiError is not null assume that the call was successful
		if resultValues[1].Interface().(*ApiError) != nil {
			//api error log error and send response
			apiError := resultValues[1].Interface().(*ApiError)
			errorType := resultValues[2].Interface().(uint8)
			color.Red("Error with method: %s, apiError: %s, errorType: %d", methodCamelCase, apiError.LogError.Error(), errorType)
			extra = apiError.UserError.Error()
			switch errorType {
			case invalidParameters:
				goto InvalidParameters
			case customError:
				goto CustomError
			}
		}
		if resultValues[0].CanInterface() {
			if resultValues[0].Interface() != nil {
				data, err = json.Marshal(resultValues[0].Interface())
				if err != nil {
					extra = "Failed to marshal content"
					goto InternalError
				}
			}
			goto Success
		}
		//if can not get interface from value there is an internal error...
		extra = fmt.Sprintf("Unable to interface successful response.")
		color.Red(extra)
		goto InternalError
	}
	return

	// Easier to handle general here
Success:
	result = json.RawMessage(data)
	resp.Result = &result
	data, _ = resp.CustomMarshalJSON()
	w.Write(data)
	return
MethodNotFound:
	w.Write(marshalErr(jsonrpc.NewMethodNotFoundError(extra, req.ID)))
	return
	//--------------
	// NEVER USED THIS... KEEP FOR NOW UNTIL WE DONT WANT IT
	//-------------
	// InvalidRequest:
	// 	w.Write(marshalErr(jsonrpc.NewInvalidRequestError(extra, req.ID)))
	// 	return
InvalidParameters:
	w.Write(marshalErr(jsonrpc.NewInvalidParametersError(extra, req.ID)))
	return
CustomError:
	w.Write(marshalErr(jsonrpc.NewCustomError(extra, req.ID, errorID)))
	return
InternalError:
	w.Write(marshalErr(jsonrpc.NewInternalRPCSError(extra, req.ID)))
	return
}

type HexInput struct {
	Hex string `json:"hex"`
}
type StringResp struct {
	Message string `json:"message"`
}

func (api *ApiProvider) LoadDay(input json.RawMessage) (successResponse *interface{}, apiError *ApiError, errorType uint8) {
	hi := new(HexInput)
	err := json.Unmarshal(input, hi)
	if err != nil {
		return nil, &ApiError{
			fmt.Errorf("Error unmarshall LoadDay: %s", err.Error()),
			fmt.Errorf("Error unmarshall LoadDay: %s", err.Error()),
		}, 0
	}

	day, err := hex.DecodeString(hi.Hex)
	if err != nil {
		return nil, &ApiError{
			fmt.Errorf("Error unmarshall LoadDay Hex: %s", err.Error()),
			fmt.Errorf("Error unmarshall LoadDay Hex: %s", err.Error()),
		}, 0
	}

	err = api.Scraper.Walker.LoadDay(day)
	if err != nil {
		return nil, &ApiError{
			fmt.Errorf("Error LoadDay: %s", err.Error()),
			fmt.Errorf("Error LoadDay: %s", err.Error()),
		}, 0
	}

	resp := new(interface{})
	*resp = &StringResp{"success"}
	return resp, nil, 0
}

/*

func (apiProvider ApiProvider) GetStats(input json.RawMessage) (successResponse *interface{}, apiError *ApiError, errorType uint8) {
	stats, err := apiProvider.Provider.GetStats()
	if err != nil {
		return nil,
			&ApiError{
				fmt.Errorf("Error retrieving stats: %s", err.Error()),
				fmt.Errorf("Error retrieving stats: %s", err.Error()),
			},
			customError
	}
	retVal := new(interface{})
	*retVal = stats
	return retVal, nil, noError
}
*/
