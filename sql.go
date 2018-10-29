package main

import (
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func setStatus(managerId string, status bool) error {
	var err error

	var updatedAt = time.Now()

	stmt, err := MySQL.Prepare("UPDATE chat_jivosite_manager_crm set online_at=? where manager_id=?")

	if err != nil {
		return err
	}

	stmt.Exec(updatedAt, managerId)

	return nil
}

func selOfflineAll() error {
	var err error

	stmt, err := MySQL.Prepare("UPDATE chat_jivosite_manager_crm set is_online=?")

	if err != nil {
		return err
	}

	stmt.Exec(false)

	return nil
}
