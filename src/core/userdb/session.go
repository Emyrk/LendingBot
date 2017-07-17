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
	DEFAULT_SESSION_DUR = time.Duration(10) * time.Minute
	SESSION_FORMAT      = "2006-01-02 15:04:05.00"
)

type Session struct {
	SessionId        string      `bson:"sessionId"`
	Email            string      `bson:"email"`
	InitialStartTime time.Time   `bson:"ist"`
	LastRenewalTime  time.Time   `bson:"lrt"`
	RenewalCount     uint64      `bson:"rc"`
	CloseTime        *time.Time  `bson:"ct"`
	CurrentIP        net.IP      `bson:"ip"`
	Open             bool        `bson:"open"`
	IPS              []SessionIP `bson:"ips"`
}

type SessionIP struct {
	IP        net.IP    `bson:"ip"`
	StartTime time.Time `bson:"st"`
}

func (ses Session) IsSameAs(sesComp *Session) bool {
	if ses.SessionId != sesComp.SessionId {
		return false
	}
	if ses.Email != sesComp.Email {
		return false
	}
	if ses.InitialStartTime.UTC().Format(SESSION_FORMAT) < sesComp.InitialStartTime.UTC().Format(SESSION_FORMAT) {
		fmt.Println("yeeeeeeeeeeess1")
		return false
	}
	if ses.LastRenewalTime.UTC().Format(SESSION_FORMAT) < sesComp.LastRenewalTime.UTC().Format(SESSION_FORMAT) {
		fmt.Println("yeeeeeeeeeeess2")
		return false
	}
	if ses.RenewalCount != sesComp.RenewalCount {
		fmt.Println("yeeeeeeeeeeess3")
		return false
	}
	if ses.CloseTime != nil {
		if sesComp.CloseTime != nil {
			if ses.CloseTime.UTC().Format(SESSION_FORMAT) < sesComp.CloseTime.UTC().Format(SESSION_FORMAT) {
				fmt.Println("yeeeeeeeeeeess4")
				return false
			}
		}
	}
	if sesComp.CloseTime != nil {
		if ses.CloseTime != nil {
			if ses.CloseTime.UTC().Format(SESSION_FORMAT) < sesComp.CloseTime.UTC().Format(SESSION_FORMAT) {
				fmt.Println("yeeeeeeeeeeess5")
				return false
			}
		}
	}
	if ses.CurrentIP.Equal(sesComp.CurrentIP) == false {
		fmt.Println("yeeeeeeeeeeess6")
		return false
	}
	if ses.Open != sesComp.Open {
		fmt.Println("yeeeeeeeeeeess7")
		return false
	}
	if len(ses.IPS) != len(sesComp.IPS) {
		fmt.Println("yeeeeeeeeeeess8")
		return false
	}
	for i, _ := range ses.IPS {
		if ses.IPS[i].StartTime.UTC().Format(SESSION_FORMAT) != sesComp.IPS[i].StartTime.UTC().Format(SESSION_FORMAT) {

			fmt.Println("yeeeeeeeeeeess10")
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

	session, err := ud.findSession(sessionId, c)
	if err != nil {
		if err.Error() == mgo.ErrNotFound.Error() {
			// if you cant find it add it
			ips := []SessionIP{SessionIP{ip, recordTime}}
			err = c.Insert(&Session{
				SessionId:        sessionId,
				Email:            email,
				InitialStartTime: recordTime,
				LastRenewalTime:  recordTime,
				RenewalCount:     0,
				CloseTime:        nil,
				CurrentIP:        ip,
				Open:             true,
				IPS:              ips,
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
	if session.Open == false {
		return fmt.Errorf("Session [%s] aready closed. Attempted update by email[%s] with ip[%s]", sessionId, email, string(ip))
	}
	// /error check

	if open == false {
		session.CloseTime = &recordTime
		session.Open = false
	} else {
		session.RenewalCount++
		session.LastRenewalTime = recordTime
	}
	if session.CurrentIP.Equal(ip) == false {
		session.CurrentIP = ip
		session.IPS = append(session.IPS, SessionIP{ip, recordTime})
	}

	//update old ones
	err = c.Update(bson.M{"sessionId": sessionId}, session)
	if err != nil {
		return err
	}
	return nil
}

func (ud *UserDatabase) findSession(sessionId string, c *mgo.Collection) (*Session, error) {
	if c == nil {
		return nil, fmt.Errorf("Error collection is nil")
	}

	var retSession Session

	find := bson.M{"sessionId": sessionId}
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
	err = c.Find(bson.M{"sessionId": sessionId}).One(&retSession)
	if err != nil {
		return nil, err
	}
	return &retSession, nil
}

func (ud *UserDatabase) CloseUserSession(sessionId string) error {
	s, c, err := ud.mdb.GetCollection(mongo.C_Session)
	if err != nil {
		return err
	}
	defer s.Close()

	//update old ones
	err = c.Update(bson.M{"sessionId": sessionId}, bson.M{"$set": bson.M{"open": false}})
	if err != nil {
		return err
	}
	return nil
}
