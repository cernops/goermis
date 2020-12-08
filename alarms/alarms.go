package alarms

import (
	"database/sql"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/miekg/dns"
	"gitlab.cern.ch/lb-experts/goermis/aiermis/orm"
)

const dnsManager = "137.138.28.176"

//CheckAlarms periodically makes sure that the thresholds are respected.
//Otherwise notifies by e-mail and updates the DB
func CheckAlarms() {
	var alarm orm.Alarm
	err := db.ManagerDB().Preload("Alarms").Find(&alarm)
	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT w.id, alias_name, name, recipient, parameter, active, last_active" +
		" FROM ermis_api_alias a  join ermis_api_alert w on  (a.id =w.alias_id) ")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	updateAlert := db.ManagerDB().Raw("UPDATE ermis_api_alert set active = ?, last_active=?, last_check =?  where id =?")
	updateNullAlert, err := db.ManagerDB().Raw("UPDATE ermis_api_alert set active = ?, last_check =?  where id =?")

	rows, err := stmtOut.Query()
	defer rows.Close()
	var myAlert Alert
	for rows.Next() {
		myAlert.LastCheck.Time = time.Now()
		err := rows.Scan(&myAlert.ID, &myAlert.Alias, &myAlert.Name, &myAlert.Recipient, &myAlert.Parameter, &myAlert.Active, &myAlert.LastActive)
		if err != nil {
			panic(err.Error())
		}
		newActive := false
		if checkAlarm(myAlert.Alias, myAlert.Name, myAlert.Parameter) {
			log.Warning("The alert should be active")
			newActive = true
			if !myAlert.Active {
				log.Info("The alert was not active before. Let's send the notification")
				myAlert.LastActive = myAlert.LastCheck
				myAlert.LastActive.Valid = true
				sendNotification(myAlert.Alias, myAlert.Recipient, myAlert.Name, myAlert.Parameter)
			}
		}

		var a sql.Result
		log.Info(fmt.Sprintf("%+v\n", myAlert))

		if myAlert.LastActive.Valid {
			a, err = updateAlert.Exec(newActive, myAlert.LastActive.Time, myAlert.LastCheck.Time, myAlert.ID)
		} else {
			a, err = updateNullAlert.Exec(newActive, myAlert.LastCheck.Time, myAlert.ID)
		}

		if err != nil {
			log.Error(fmt.Sprintf("Error updating the alert: %v and %v", err, a))
		}
	}
}

func sendNotification(alias, recipient, name string, parameter int) {
	log.Info(fmt.Sprintf("Sending a notification to %v that the alert %s on %s has been triggered (less than %d nodes)", recipient, alias, name, parameter))
	msg := []byte("To: " + alias + "\r\n" +
		fmt.Sprintf("Subject: Alert on the alias %s: only %d hosts\r\n\r\nThe alert %s (%d) on %s has been triggered", alias, parameter, name, parameter, alias))
	err := smtp.SendMail("localhost:25", nil, "kristian.kouros@cern.ch", []string{recipient}, msg)
	if err != nil {
		log.Error(err)
	}

}

func checkAlarm(alias, alert string, parameter int) bool {
	log.Info(fmt.Sprintf("Checking if the alarm %s %v on %s is active", alert, parameter, alias))
	if alert == "minimum" {
		return checkMinimumAlarm(alias, parameter)
	}
	log.Error(fmt.Sprintf("The alert %v (on %v) is not understood!", alert, alias))
	return true
}

func getIpsFromDNS(m *dns.Msg, alias, dnsManager string, dnsType uint16, ips *[]net.IP) error {
	m.SetQuestion(alias+".", dnsType)
	in, err := dns.Exchange(m, dnsManager+":53")
	if err != nil {
		log.Error(fmt.Sprintf("Error getting the ipv4 state of dns: %v", err))
		return err
	}
	for _, a := range in.Answer {
		if t, ok := a.(*dns.A); ok {
			log.Debug(fmt.Sprintf("From %v, got ipv4 %v", t, t.A))
			*ips = append(*ips, t.A)
		} else if t, ok := a.(*dns.AAAA); ok {
			log.Debug(fmt.Sprintf("From %v, got ipv6 %v", t, t.AAAA))
			*ips = append(*ips, t.AAAA)
		}
	}
	return nil
}
func checkMinimumAlarm(alias string, parameter int) bool {
	m := new(dns.Msg)
	var ips []net.IP
	m.SetEdns0(4096, false)
	log.Info("Getting the ips from the DNS")
	err := getIpsFromDNS(m, alias, dnsManager, dns.TypeA, &ips)

	if err != nil {
		return true
	}
	err = getIpsFromDNS(m, alias, dnsManager, dns.TypeAAAA, &ips)
	if err != nil {
		return true
	}
	log.Info(fmt.Sprintf("The list of ips : %v\n", ips))
	if len(ips) < parameter {
		log.Info(fmt.Sprintf("There are less than %d nodes (only %d)\n", parameter, len(ips)))
		return true
	}

	return false
}
