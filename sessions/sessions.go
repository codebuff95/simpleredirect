package sessions

/*
Sessions are used in this project to authenticate form and user sessions through their
SessionIDs (SIDs). This makes sure that no form is submitted multiple times, no form is submitted
after it has been expured. Similarly, no user can remain logged in after his/her UserSessionID
expires.
*/
import (
	"log"
	"math/rand" //preferable replacement: "crypto/rand".
	"simpleredirect/databases"
	"strings"
	"time"
)

//GlobalSM is a map that contains all the SessionManagers. Used map for easy and unambiguous access
//to the SessionManagers.
var GlobalSM map[string]*SessionManager

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

const (
	EXPIRED int = 1
	ACTIVE  int = 2
	DELETED int = 3
	//SIDLEN Length of each SessionID.
	SIDLEN int = 16
)

type Session struct {
	Sid    string //unique sessionid
	Rid    string //required id that corresponds to the sid
	Status int    //status of session. eg- expired, active, deleted, etc.
}

type SessionManager struct {
	Db        *databases.DBManager
	TableName string
}

func InitGlobalSM() {
	GlobalSM = make(map[string]*SessionManager)
}

func GenerateUniqueSid() string {
	b := make([]rune, SIDLEN)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Authenticate authenticates the sid of SessionManager sm, and returns the corresponding Rid if valid session, else
// returns empty string.
func (sm *SessionManager) Authenticate(sid string) string {
	mySession := sm.GetSession(sid)
	if mySession != nil {
		log.Println("session status is", mySession.Status)
	} else {
		log.Println("session returned is nil.")
	}
	if mySession != nil && mySession.Status == ACTIVE {
		return mySession.Rid
	}
	log.Println("Invalid Sid:", sid)
	return ""
}

func (sm *SessionManager) GetSession(sid string) *Session {
	log.Println("checking for sid:", sid)
	row := sm.Db.Con.QueryRow("SELECT * FROM " + sm.TableName + " WHERE sid = '" + sid + "'")
	mysession := Session{}
	var expires string
	err := row.Scan(&mysession.Sid, &mysession.Rid, &expires)
	if err != nil {
		return &Session{Status: DELETED} //or does not exist.
	}
	timenow := time.Now().Format("2006-01-02 15:04:05")
	if i := strings.Compare(expires, timenow); i <= 0 {
		sm.DeleteSession(sid)
		return &Session{Status: EXPIRED} // and has been deleted due to expiry.
	} else {
		mysession.Status = ACTIVE
		return &mysession
	}
}

func (sm *SessionManager) SetSession(rid string, life time.Duration) *Session {
	sid := GenerateUniqueSid()
	// delete all sessions in sm.TableName table with rid = 'rid'
	/*fmt.Println("Delete from table where rid??")
	stmt, err := sm.Db.Con.Prepare("DELETE FROM ? WHERE rid = '?'")
	if err != nil{
		log.Println(err)
		return nil
	}
	fmt.Println("Deleted?? from table where rid??")
	res,err := stmt.Exec(sm.TableName, rid)
	rowsdeleted, _ := res.RowsAffected()
	log.Println("deleted",rowsdeleted,"rows for rid =",rid)*/ // No Need to Delete all sids with this rid.

	//Create Session
	expiry := time.Now().Add(life).Format("2006-01-02 15:04:05")
	//stmt, err := sm.Db.Con.Prepare("INSERT INTO ? SET sid = '?', rid = '?', expires = '?'")
	stmt, err := sm.Db.Con.Prepare("INSERT INTO " + sm.TableName + " SET sid = '" + sid + "', rid = '" + rid + "', expires = '" + expiry + "'")
	if err != nil {
		log.Println(err)
		return nil
	}
	//_,err = stmt.Exec(sm.TableName,sid,rid,expiry)
	_, err = stmt.Exec()
	if err != nil {
		log.Println(err)
		return nil
	}
	return &Session{Sid: sid, Rid: rid, Status: ACTIVE}
}

func (sm *SessionManager) DeleteSession(sid string) int64 {
	log.Println("Deleting Session with sid:", sid, "tablename:", sm.TableName)
	stmt, err := sm.Db.Con.Prepare("DELETE FROM " + sm.TableName + " WHERE sid = '" + sid + "'")
	if err != nil {
		log.Println(err)
		return int64(0)
	}
	res, err := stmt.Exec()
	rowsdeleted, _ := res.RowsAffected()
	log.Println("deleted", rowsdeleted, "rows for sid =", sid)
	return rowsdeleted
}
