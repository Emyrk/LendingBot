package userdb

import (
	"fmt"
	"net"
	"time"

	"github.com/Emyrk/LendingBot/src/core/database/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DEFAULT_SESSION_DUR = 5 * time.Minute //should use a const value, should be same as the default in session look at session.go in controllers
	SESSION_FORMAT      = "2006-01-02 15:04:05.000"
)

type Session struct {
	SessionId       string         `json:"sessionId" bson:"sessionId"`
	Email           string         `json:"email" bson:"email"`
	LastRenewalTime time.Time      `json:"lrt" bson:"lrt"`
	CurrentIP       net.IP         `json:"ip" bson:"ip"`
	Open            bool           `json:"open" bson:"open"`
	IPS             []SessionIP    `json:"ips" bson:"ips"`
	ChangeState     []SessionState `json:"changestate" bson:"changestate"` //tracks when session has changed from opened to close or when reopened
}

type SessionIP struct {
	IP        net.IP    `json:"ip" bson:"ip"`
	StartTime time.Time `json:"st" bson:"st"`
}

type SessionAction string

const (
	OPENED     SessionAction = "O"
	CLOSED     SessionAction = "C"
	REOPENED   SessionAction = "R"  //occurs with timeout of session so reloging in
	INITCLOSED SessionAction = "IC" //occurs on start up. Will have to change later. Time will be incorrect
)

type SessionState struct {
	SessionAction SessionAction `bson:"sessionaction"`
	ActionTime    time.Time     `bson:"actiontime,omitempty"`
}

func (ses Session) IsSameAs(sesComp *Session) bool {
	if ses.SessionId != sesComp.SessionId {
		return false
	}
	if ses.Email != sesComp.Email {
		return false
	}
	if ses.LastRenewalTime.UTC().Format(SESSION_FORMAT) < sesComp.LastRenewalTime.UTC().Format(SESSION_FORMAT) {
		return false
	}
	if ses.CurrentIP.Equal(sesComp.CurrentIP) == false {
		return false
	}
	if ses.Open != sesComp.Open {
		return false
	}
	if len(ses.IPS) != len(sesComp.IPS) {
		return false
	}
	for i, _ := range ses.IPS {
		if ses.IPS[i].StartTime.UTC().Format(SESSION_FORMAT) != sesComp.IPS[i].StartTime.UTC().Format(SESSION_FORMAT) {
			return false
		}
	}
	if len(ses.ChangeState) != len(sesComp.ChangeState) {
		return false
	}
	for i, _ := range ses.ChangeState {
		if ses.ChangeState[i].ActionTime.UTC().Format(SESSION_FORMAT) != sesComp.ChangeState[i].ActionTime.UTC().Format(SESSION_FORMAT) {
			return false
		}
		if ses.ChangeState[i].SessionAction != sesComp.ChangeState[i].SessionAction {
			return false
		}
	}
	return true
}

func (ud *UserDatabase) UpdateUserSession(sessionId, email string, recordTime time.Time, ip net.IP, open bool) error {
	s, c, err := ud.mdb.GetCollection(mongo.C_Session)
	if err != nil {
		return err
	}
	defer s.Close()

	recordTime = recordTime.UTC()

	session, err := ud.findSession(sessionId, email, c)
	if err != nil {
		if err.Error() == mgo.ErrNotFound.Error() {
			// if you cant find it add it
			ips := []SessionIP{SessionIP{ip, recordTime}}
			cs := []SessionState{SessionState{OPENED, recordTime}}
			err = c.Insert(&Session{
				SessionId:       sessionId,
				Email:           email,
				LastRenewalTime: recordTime,
				CurrentIP:       ip,
				Open:            true,
				IPS:             ips,
				ChangeState:     cs,
			})
			if err != nil {
				return err
			}
			return nil
		}
		// if another error return that error
		return err
	}

	//compare email to session
	//error check
	if session.Email != email {
		return fmt.Errorf("Emails do not match session[%s]. Session email [%s]. Given email [%s]", sessionId, session.Email, email)
	}
	if session.Open == false && open == false {
		return fmt.Errorf("Session [%s] aready closed, trying to close sesssion again. Attempted update by email[%s] with ip[%s]", sessionId, email, ip.String())
	}
	// /error check

	push := bson.M{}
	if session.Open == false && open == true {
		push["changestate"] = SessionState{REOPENED, recordTime}
		session.Open = true
		session.LastRenewalTime = recordTime
	} else if session.Open == true && open == false {
		push["changestate"] = SessionState{CLOSED, recordTime}
		session.Open = false
	} else if session.Open == true && open == true {
		session.LastRenewalTime = recordTime
	}

	if session.CurrentIP.Equal(ip) == false {
		session.CurrentIP = ip
		push["ips"] = SessionIP{ip, recordTime}
	}

	//bson.M{"changestate": newSessionState, "ips": newIP}

	update := bson.M{
		"$set": bson.M{
			"open": open,
			"lrt":  session.LastRenewalTime,
			"ip":   ip,
		},
	}
	if len(push) != 0 {
		update["$push"] = push
	}
	//update old ones
	err = c.Update(bson.M{"sessionId": sessionId, "email": email}, update)
	if err != nil {
		return err
	}
	return nil
}

func (ud *UserDatabase) findSession(sessionId, email string, c *mgo.Collection) (*Session, error) {
	if c == nil {
		return nil, fmt.Errorf("Error collection is nil")
	}

	var retSession Session

	find := bson.M{"sessionId": sessionId, "email": email}
	err := c.Find(find).One(&retSession)
	if err != nil {
		return nil, err
	}
	return &retSession, nil
}

// 0: both open and closed sessions
// 1: only open sessions
// 2: only closed sessions
func (ud *UserDatabase) GetAllUserSessions(email string, open uint8, limit int) (*[]Session, error) {
	s, c, err := ud.mdb.GetCollection(mongo.C_Session)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	var retSessions []Session
	find := bson.M{"email": email}
	if open == 1 {
		find = bson.M{"email": email, "open": true}
	} else if open == 2 {
		find = bson.M{"email": email, "open": false}
	}
	err = c.Find(find).Sort("-lrt").Limit(limit).All(&retSessions)
	if err != nil {
		return nil, err
	}
	return &retSessions, nil
}

func (ud *UserDatabase) GetUserSession(sessionId, email string) (*Session, error) {
	s, c, err := ud.mdb.GetCollection(mongo.C_Session)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	var retSession Session
	err = c.Find(bson.M{"sessionId": sessionId, "email": email}).One(&retSession)
	if err != nil {
		return nil, err
	}
	return &retSession, nil
}

func (ud *UserDatabase) CloseUserSession(sessionId, email string) error {
	s, c, err := ud.mdb.GetCollection(mongo.C_Session)
	if err != nil {
		return err
	}
	defer s.Close()

	update := bson.M{
		"$set":  bson.M{"open": false},
		"$push": bson.M{"changestate": SessionState{CLOSED, time.Now().UTC()}},
	}
	//update old ones
	err = c.Update(bson.M{"sessionId": sessionId, "email": email}, update)
	if err != nil {
		fmt.Println("failed", err.Error())
		return err
	}
	fmt.Println("ADDED")
	return nil
}

func (ud *UserDatabase) CloseAllSessions() error {
	s, c, err := ud.mdb.GetCollection(mongo.C_Session)
	if err != nil {
		return err
	}
	defer s.Close()

	update := bson.M{
		"$set":  bson.M{"open": false},
		"$push": bson.M{"changestate": SessionState{INITCLOSED, time.Now().UTC()}},
	}
	//update old ones
	err = c.Update(bson.M{}, update)
	if err != nil {
		fmt.Println("failed", err.Error())
		return err
	}
	return nil
}
