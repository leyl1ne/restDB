package restdb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"

	_ "github.com/lib/pq"
)

type User struct {
	ID        int
	Username  string
	Password  string
	LastLogin int64
	Admin     int
	Active    int
}

var Hostname = "localhost"
var Port = 5432
var Username = "leyline"
var Password = "pass"
var Database = "restapi"

func (p *User) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(p)
}

func (p *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func ConnectPostgres() *sql.DB {
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Println(err)
		return nil
	}

	return db
}

func DeleteUser(ID int) bool {
	db := ConnectPostgres()
	if db == nil {
		log.Println("Cannot connect to PostgreSQL")
		db.Close()
		return false
	}
	defer db.Close()

	t := FindUserID(ID)
	if t.ID == 0 {
		log.Println("User", ID, "does not exist.")
		return false
	}

	stmt, err := db.Prepare("DELETE FROM users WHERE ID = $1")
	if err != nil {
		log.Println("DeleteUser:", err)
		return false
	}

	_, err = stmt.Exec(ID)
	if err != nil {
		log.Println("DeleteUser:", err)
		return false
	}

	return true
}

func InserUser(u User) bool {
	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect ot PostgreSQL!")
		return false
	}
	defer db.Close()

	if IsUserValid(u) {
		log.Println("User", u.Username, "already exists!")
		return false
	}

	stmt, err := db.Prepare("INSERT INTO user(Username, Password, LastLogin, Admin, Active) values($1,$2,$3,$4,$5)")
	if err != nil {
		log.Println("Add user:", err)
		return false
	}

	stmt.Exec(u.Username, u.Password, u.LastLogin, u.Admin, u.Active)
	return true
}

func ListAllUsers() []User {
	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect to PostgreSQL!")
		db.Close()
		return []User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users \n")
	if err != nil {
		log.Println(err)
		return []User{}
	}

	all := []User{}
	var id int
	var username, password string
	var lastLogin int64
	var admin, active int

	for rows.Next() {
		err = rows.Scan(&id, &username, &password, &lastLogin, &admin, &active)
		temp := User{id, username, password, lastLogin, admin, active}
		all = append(all, temp)
	}

	log.Println("All:", all)
	return all
}

func FindUserID(ID int) User {
	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect to PostgreSQL!")
		db.Close()
		return User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FORM users WHERE ID = $1\n", ID)
	if err != nil {
		log.Println("Query:", err)
		return User{}
	}
	defer rows.Close()

	u := User{}
	var id int
	var username, password string
	var lastLogin int64
	var admin, active int

	for rows.Next() {
		err = rows.Scan(&id, &username, &password, &lastLogin, &admin, &active)
		if err != nil {
			log.Println(err)
			return User{}
		}
		u = User{id, username, password, lastLogin, admin, active}
		log.Println("Found user:", u)
	}
	return u
}

func FindUserUsername(usernameQuery string) User {
	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect to PostgreSQL!")
		db.Close()
		return User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE Username = $1 \n", usernameQuery)
	if err != nil {
		log.Println("FindUserUsername Query:", err)
		return User{}
	}
	defer rows.Close()

	u := User{}
	var id int
	var username, password string
	var lastLogin int64
	var admin, active int

	for rows.Next() {
		err = rows.Scan(&id, &username, &password, &lastLogin, &admin, &active)
		if err != nil {
			log.Println(err)
			return User{}
		}
		u = User{id, username, password, lastLogin, admin, active}
		log.Println("Found user:", u)
	}
	return u
}

func ReturnLoggedUsers() []User {
	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect to PostgreSQL!")
		db.Close()
		return []User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE Active = 1 \n")
	if err != nil {
		log.Println(err)
		return []User{}
	}

	all := []User{}
	var id int
	var username, password string
	var lastLogin int64
	var admin, active int

	for rows.Next() {
		err = rows.Scan(&id, &username, &password, &lastLogin, &admin, &active)
		if err != nil {
			log.Println(err)
			return []User{}
		}
		temp := User{id, username, password, lastLogin, admin, active}
		log.Println("temp:", all)
		all = append(all, temp)
	}

	log.Println("Logged in:", all)
	return all
}

func IsUserAdmin(u User) bool {
	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect to PostgreSQL!")
		db.Close()
		return false
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE Username = $1 \n", u.Username)
	if err != nil {
		log.Println(err)
		return false
	}

	temp := User{}
	var id int
	var username, password string
	var lastLogin int64
	var admin, active int

	// we will get the FIRST ONE only
	for rows.Next() {
		err = rows.Scan(&id, &username, &password, &lastLogin, &admin, &active)
		if err != nil {
			log.Println(err)
			return false
		}
		temp = User{id, username, password, lastLogin, admin, active}
	}

	if u.Username == temp.Username && u.Password == temp.Password && temp.Admin == 1 {
		return true
	}
	return false
}

func UpdateUser(u User) bool {
	log.Println("Updating user:", u)

	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect to PostgreSQL!")
		db.Close()
		return false
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE users SET Username=$1, Password=$2, Admin=$3, Active=$4 WHERE ID = $5")
	if err != nil {
		log.Println("UpdateUser prepare:", err)
		return false
	}

	res, err := stmt.Exec(u.Username, u.Password, u.Admin, u.Active, u.ID)
	if err != nil {
		log.Println("UpdateUser failed:", err)
		return false
	}

	//checking the changes made
	affect, err := res.RowsAffected()
	if err != nil {
		log.Println("RowsAffected() failed:", err)
		return false
	}
	log.Println("Affected:", affect)
	return true
}

func IsUserValid(u User) bool {
	db := ConnectPostgres()
	if db == nil {
		fmt.Println("Cannot connect to PostgreSQL!")
		db.Close()
		return false
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE Username = $1 \n", u.Username)
	if err != nil {
		log.Println(err)
		return false
	}

	temp := User{}
	var id int
	var username, password string
	var lastLogin int64
	var admin, active int

	// if exist several users by the same username
	// then we will get only the first one
	for rows.Next() {
		err = rows.Scan(&id, &username, &password, &lastLogin, &admin, &active)
		if err != nil {
			log.Println(err)
			return false
		}
		temp = User{id, username, password, lastLogin, admin, active}
	}

	if u.Username == temp.Username && u.Password == temp.Password {
		return true
	}

	return false
}
