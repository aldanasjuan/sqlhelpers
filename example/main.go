package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aldanasjuan/sqlhelpers"
)

func main() {
	start := time.Now()
	f, _ := sqlhelpers.StructMap(User{})
	_ = f
	res, newMap, err := sqlhelpers.Migrate(f, User2{}, "users", false)
	if err != nil {
		fmt.Println("error", err)
	}
	for _, r := range res {
		fmt.Println(r)
	}
	//save this map in the db
	js, _ := json.MarshalIndent(newMap, "", "  ")
	fmt.Println(string(js))
	fmt.Println(time.Since(start))

}

type User struct {
	ID                    int     `json:"id,omitempty" db:"field:bigserial not null primary key"`
	Node                  string  `json:"node,omitempty" db:"field:text not null"`
	FirstName             string  `json:"first_name,omitempty" db:"field:text not null"`
	LastName              string  `json:"last_name,omitempty" db:"field:text not null"`
	Email                 string  `json:"email,omitempty" db:"field:text not null unique"`
	ChargeBeeID           string  `json:"charge_bee_id,omitempty" db:"field:text not null"`
	ChargeBeeSubscription string  `json:"charge_bee_subscription,omitempty" db:"field:text not null"`
	TrialStart            int     `json:"trial_start,omitempty" db:"field:int"`
	TrialEnd              int     `json:"trial_end,omitempty" db:"field:int check(id > 2)"`
	ResetToken            *string `json:"token,omitempty" db:"field:text default('bye')"`
	SomethingID           int     `json:"something_id,omitempty" db:"field:bigint references something(id) on delete cascade on update set null"`
}

type User2 struct {
	ID                    int     `json:"id,omitempty" db:"field:bigserial not null primary key"`
	Node                  string  `json:"node,omitempty" db:"field:text not null"`
	FirstName             string  `json:"first_name,omitempty" db:"field:text not null"`
	LastName              string  `json:"last_name,omitempty" db:"field:text not null"`
	Email                 string  `json:"email,omitempty" db:"field:text not null unique"`
	ChargeBeeID           string  `json:"charge_bee_id,omitempty" db:"field:text not null"`
	ChargeBeeSubscription string  `json:"charge_bee_subscription,omitempty" db:"field:text not null"`
	TrialStart            int     `json:"trial_start,omitempty" db:"field:int"`
	TrialEnd              int     `json:"trial_end,omitempty" db:"field:int check(id > 2)"`
	ResetToken            *string `json:"token,omitempty" db:"field:text default('bye')"`
	SomethingID           int     `json:"something_id,omitempty" db:"field:text references something(id) on delete cascade on update set null"`
}
