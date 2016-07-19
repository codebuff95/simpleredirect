package databases

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//GlobalDBM is a map that contains all the DatabaseManagers. Used map for easy and unambiguous access
//to the DBManagers.
var GlobalDBM map[string]*DBManager

//DBManager type contains all the necessary details of a DBManager.s
type DBManager struct {
	Name     string
	Database string
	User     string
	Password string
	Con      *sql.DB
}

//InitGlobalDBM initialises the GlobalDBM map.
func InitGlobalDBM() {
	GlobalDBM = make(map[string]*DBManager)
}

//Open opens a database connection on the dbm.Con field.
func (dbm *DBManager) Open() error {
	var err error
	dbm.Con, err = sql.Open(dbm.Name, dbm.User+":"+dbm.Password+"@/"+dbm.Database)
	return err
}

var cleanInXMinutes time.Duration = 60

//CleanTable deletes all entries from the table 'tablename' of DatabaseManager which have expired.
//It can be used to clean tables like formsession and usersession.
//Cleaning takes place at an interval of cleanInXMinutes minutes (adjust this variable according
// to the site's traffic).
func (dbm *DBManager) CleanTable(tablename string) {
	log.Println("Initialising Cleantable on table", tablename, "of databasemanager", dbm.Name)
	for {
		nowTime := time.Now().Format("2006-01-02 15:04:05")
		stmt, err := dbm.Con.Prepare("DELETE FROM " + tablename + " WHERE STRCMP(expires,'" + nowTime + "') = -1")
		if err != nil {
			log.Println("Could not initiate TableCleaner on table", tablename, "of databasemanager", dbm.Name)
			log.Println(err)
			return
		}
		res, err := stmt.Exec()
		log.Println("** Status of Cleantable on table", tablename, "of databasemanager", dbm.Name, "**")
		if err != nil {
			log.Println("deleted 0 rows.")
		} else {
			rowsdeleted, _ := res.RowsAffected()
			log.Println("deleted", rowsdeleted, "rows.")
		}
		time.Sleep(time.Minute * cleanInXMinutes) // Cleaning takes place every 60 minutes.
	}
}
