package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	_ "github.com/go-sql-driver/mysql"
)

type Host struct {
	Url    string
	Status sql.NullString
}

var checked_hosts []Host

func check_host(host Host, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := http.Get(host.Url)
	if err != nil {
		host.Status = sql.NullString{String: "0"}
	} else {
		fmt.Print(strconv.Itoa(resp.StatusCode) + " : ")
		fmt.Println(host.Url)
		if resp.StatusCode == 200 {
			host.Status = sql.NullString{String: "1"}
		} else {
			host.Status = sql.NullString{String: "0"}
		}
	}
	checked_hosts = append(checked_hosts, host)
}

func main() {
	log.Println("...host_checker... started")

	db, err := sql.Open("sqlite3", "./host_checker.db")
	// db, err := sql.Open("mysql", "user7:s$cret@tcp(127.0.0.1:3306)/testdb")
	checkErr(err)
	log.Println("Connected to database!")
	log.Println("Getting rows from database...")
	rows, err := db.Query("SELECT url, status FROM hosts")
	checkErr(err)

	var hosts []Host
	for rows.Next() {
		var h Host
		err := rows.Scan(&h.Url, &h.Status)
		checkErr(err)

		hosts = append(hosts, h)
	}

	log.Println("Got " + strconv.Itoa(len(hosts)) + " rows from database!")
	fmt.Println("-------------------------------------")

	start := time.Now()
	var wg sync.WaitGroup
	log.Println("Checking hosts...")
	for _, host := range hosts {
		wg.Add(1)
		go check_host(host, &wg)
	}

	wg.Wait()
	elapsed := time.Since(start)

	log.Println("It took " + elapsed.String() + " time to check all hosts!")
	log.Println("Saving checked hosts to database...")

	stmt, err := db.Prepare("update hosts set status=? where url=?")
	checkErr(err)

	for _, checked_host := range checked_hosts {
		_, err = stmt.Exec(checked_host.Status.String, checked_host.Url)
		checkErr(err)
	}

	log.Println("All checked hosts are saved to database!")

	db.Close()
	log.Println("Database connection closed!")

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
