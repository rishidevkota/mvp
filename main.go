//Rishi Ram Devkota
//2017 Dec
package main
import (
	"net/http"
	"html/template"
	"github.com/rishidevkota/mvp/db"
	"strconv"
)
type Account struct {
	Guid int
	Name string
	AccountType string
	ParentGuid int
	Placeholder bool
}
type Transaction struct {
	Guid int
	Date string
	Num string
	Description string
	Debit int
	Credit int
}
type Pathelement struct{
	Guid int
	Name string
}
var accountTemplate = template.Must(template.ParseFiles("accounts.html"))

func accounts(w http.ResponseWriter, r *http.Request) {
	/*TODO:
	path
	WITH tblParent AS
				(
    				SELECT *
        			FROM accounts WHERE guid = ?
    				UNION ALL
    				SELECT accounts.*
        			FROM accounts  JOIN tblParent  ON accounts.guid = tblParent.parent_guid
				)
				SELECT guid, name, placeholder FROM  tblParent where name <> 'ROOT' order by guid`
	*/
	guid := r.FormValue("guid")
	var account Account

	switch r.Method {
	case "GET":
		if guid != "" {
			err := db.QueryRow("select guid, account_type, placeholder from accounts where guid=?", guid).Scan(&account.Guid,
				&account.AccountType, &account.Placeholder)
			if err!= nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var path []Pathelement
			rows := db.Query(`WITH parent AS
				(
    				SELECT *
        			FROM accounts WHERE guid = ?
    				UNION ALL
    				SELECT accounts.*
        			FROM accounts  JOIN parent  ON accounts.guid = parent.parent_guid
				)
				SELECT guid, name FROM  parent order by guid`, guid)
			for rows.Next() {
				var pathelement Pathelement
				err := rows.Scan(&pathelement.Guid, &pathelement.Name)
				if err!= nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				path = append(path, pathelement)
			}

			var childaccounts []Account
			var transactions []Transaction
			if account.Placeholder {
				rows := db.Query("select guid, name from accounts where parent_guid=?", guid)
				for rows.Next() {
					var childaccount Account
					err := rows.Scan(&childaccount.Guid, &childaccount.Name)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					childaccounts = append(childaccounts, childaccount)
				}
			} else {
				rows := db.Query(`SELECT t.guid, t.date, t.num, t.description, s.value
					FROM transactions t
					LEFT JOIN splits s ON t.guid = s.tx_guid
					WHERE s.account_guid = ?`, guid)
				for rows.Next() {
					var transaction Transaction
					var value int
					err := rows.Scan(
						&transaction.Guid,
						&transaction.Date,
						&transaction.Num,
						&transaction.Description,
						&value,
					)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					if value > 0 {
						transaction.Debit = value
					} else if value < 0 {
						transaction.Credit = value*(-1)
					}
					transactions = append(transactions, transaction)

				} 
			}
			accountTemplate.Execute(w, map[string]interface{}{
				"Account": account,
				"Path": path,
				"Childaccounts": childaccounts,
				"Transactions": transactions,
			})
			
		} else {
			db.QueryRow("select guid from accounts where account_type=?", "ROOT").Scan(&account.Guid)
			http.Redirect(w, r, "/accounts?guid="+strconv.Itoa(account.Guid), http.StatusFound)
		}
	case "POST":
		var placeholder bool
		if r.FormValue("placeholder") == "placeholder" {
			placeholder = true
		}
		db.Exec(`insert into accounts(
			name, code, description, account_type, placeholder, parent_guid
			) values(?, ?, ?, ?, ?,?)`,
			r.FormValue("name"),
			r.FormValue("code"),
			r.FormValue("description"),
			r.FormValue("account_type"),
			placeholder,
			r.FormValue("parent_guid"))
	default:
		w.Write([]byte("implement me"))
	}
}

func transaction(w http.ResponseWriter, r *http.Request) {
	/*TODO:
	INSERT INTO table3 ( name, age, sex, city, id, number, nationality)
	SELECT name, age, sex, city, p.id, number, n.nationality
	FROM table1 p
	INNER JOIN table2 c ON c.Id = p.Id
	*/
	debit, _ := strconv.Atoi(r.FormValue("debit"))
	credit, _ := strconv.Atoi(r.FormValue("credit"))
	value := debit - credit

	result := db.Exec(`insert into transactions(num, date, description) values(?, ?, ?)`,
		r.FormValue("num"),
		r.FormValue("date"),
		r.FormValue("description"),
	)
	tx_guid, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	db.Exec(`insert into splits(tx_guid, account_guid, value) values(?,?,?)`,
		tx_guid,
		r.FormValue("mguid"),
		value)
	db.Exec(`insert into splits(tx_guid, account_guid, value) values(?,?,?)`,
		tx_guid,
		r.FormValue("transfer"),
		-value)
}
func main() {
	http.HandleFunc("/accounts", accounts)
	http.HandleFunc("/transaction", transaction)
	http.ListenAndServe(":8080", nil)
}