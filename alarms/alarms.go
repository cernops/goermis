package alarms

/*This file periodically checks the alarms. If needed,
updates them in DB and notifies user*/

import (
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/miekg/dns"
	"gitlab.cern.ch/lb-experts/goermis/api/ermis"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
	"gitlab.cern.ch/lb-experts/goermis/db"
)

var (
	log        = bootstrap.GetLog()
	dnsManager = bootstrap.GetConf().DNS.Manager
)

//PeriodicAlarmCheck periodically makes sure that the thresholds are respected.
//Otherwise notifies by e-mail and updates the DB
func PeriodicAlarmCheck() {
	var alarms []ermis.Alarm
	if err := db.GetConn().Find(&alarms).
		Error; err != nil {
		log.Error("Could not retrieve alarms", err.Error())
	}

	for _, alarm := range alarms {
		if err := processThis(alarm); err != nil {
			log.Error(fmt.Sprintf("Error updating the alert: %v and %v", err, alarm))
		}
	}
}

func processThis(alarm ermis.Alarm) (err error) {
	alarm.LastCheck.Time = time.Now()
	alarm.LastCheck.Valid = true
	newActive := false
	if checkAlarm(alarm.Alias, alarm.Name, alarm.Parameter) {
		log.Warn("The alert should be active")
		newActive = true
		if !alarm.Active {
			log.Info("The alert was not active before. Let's send the notification")
			alarm.LastActive = alarm.LastCheck
			alarm.LastActive.Valid = true
			err := SendNotification(alarm.Alias, alarm.Recipient, alarm.Name, alarm.Parameter)
			if err != nil {
				log.Error(err)
			}
		}
	}

	if alarm.LastActive.Valid {
		err = db.GetConn().Model(&alarm).Updates(ermis.Alarm{
			Active:     newActive,
			LastActive: alarm.LastActive,
			LastCheck:  alarm.LastCheck}).Error
	} else {
		err = db.GetConn().Model(&alarm).Updates(ermis.Alarm{
			Active:    newActive,
			LastCheck: alarm.LastCheck}).Error
	}
	return err
}

//SendNotification sends an e-mail to the recipient when alarm is triggered
func SendNotification(alias, recipient, name string, parameter int) error {
	log.Info(fmt.Sprintf("Sending a notification to %v that the alert %s on %s has been triggered (less than %d nodes)", recipient, alias, name, parameter))
	msg := []byte("To: " + alias + "\r\n" +
		fmt.Sprintf("Subject: Alert on the alias %s: only %d hosts\r\n\r\nThe alert %s (%d) on %s has been triggered", alias, parameter, name, parameter, alias))
	err := smtp.SendMail("localhost:25",
		nil,
		"lbd@cern.ch",
		[]string{recipient},
		msg)
	return err

}

func checkAlarm(alias, alert string, parameter int) bool {
	log.Info(fmt.Sprintf("Checking if the alarm %s %v on %s is active", alert, parameter, alias))
	if alert == "minimum" {
		return CheckMinimumAlarm(alias, parameter)
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

//CheckMinimumAlarm compares the threshold parameter with the number of nodes behind and alias
func CheckMinimumAlarm(alias string, parameter int) bool {
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
