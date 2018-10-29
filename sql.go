package main

import (
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func setStatus(manager_id string, status bool) error {
	var err error

	var updated_at = time.Now()

	stmt, err := MySQL.Prepare("UPDATE chat_jivosite_manager_crm set is_online=?, updated_at=? where manager_id=?")

	if err != nil {
		return err
	}

	stmt.Exec(status, updated_at, manager_id)

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
