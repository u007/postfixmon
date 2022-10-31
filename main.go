package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"postfixmon/tools"
	"postfixmon/virtualmin"
	"postfixmon/whm"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var configPath = ".config"
var dataPath = "data/"

// date, id, <=, email, extras
/*
Oct 22 17:48:40 vmi673390 postfix/trivial-rewrite[183180]: warning: do not list domain s12.mercstudio.com in BOTH mydestination and virtual_alias_domains
Oct 22 17:48:40 vmi673390 postfix/smtpd[183177]: 3B838E03CA: client=localhost[127.0.0.1]
Oct 22 17:48:40 vmi673390 postfix/cleanup[183181]: 3B838E03CA: message-id=<1666432120.183174@s12.mercstudio.com>
Oct 22 17:48:40 vmi673390 postfix/qmgr[183159]: 3B838E03CA: from=<james@s12.mercstudio.com>, size=688, nrcpt=2 (queue active)
Oct 22 17:48:40 vmi673390 postfix/smtpd[183177]: disconnect from localhost[127.0.0.1] helo=1 mail=1 rcpt=2 data=1 quit=1 commands=6
Oct 22 17:48:42 vmi673390 postfix/smtp[183182]: 3B838E03CA: to=<c00lways@gmail.com>, relay=mail.smtp2go.com[45.79.71.155]:2525, delay=2.3, delays=0.05/0.01/1.7/0.62, dsn=2.0.0, status=sent (250 OK id=1omB7Q-DvC2hB-2l)
Oct 22 17:48:42 vmi673390 postfix/smtp[183182]: 3B838E03CA: to=<james@mercstudio.com>, relay=mail.smtp2go.com[45.79.71.155]:2525, delay=2.3, delays=0.05/0.01/1.7/0.62, dsn=2.0.0, status=sent (250 OK id=1omB7Q-DvC2hB-2l)

*/
// var postfixRegLine = regexp.MustCompile("(?i)([a-z]* \\d+ \\d+:\\d+:\\d+) [a-zA-Z0-9_]* postfix/[a-z]*\\[\\d*\\]: ([a-z0-9]*):.*$")
var postfixRegLine = regexp.MustCompile("(?i)([a-z]* \\d+ \\d+:\\d+:\\d+) [a-zA-Z0-9_]* postfix/[a-z]*\\[\\d*\\]: ([a-z0-9]*): (from|to)=<([a-z0-9._%+\\-]+@[a-z0-9.\\-]+\\.[a-z]{2,4})>,.*$")
// var postfixRegLine = regexp.MustCompile("(?i)(\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}) ([^ ]*) ([^ ]*) .* A=dovecot_[a-zA-z]*:([^ ]*) (.*) for (.*)$")
var notifyEmail = ""

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return defaultValue
    }
    return value
}


func main() {
	logFile := getEnv("PF_LOG", "/var/log/mail.log")
	serverType := getEnv("SERVERTYPE", "virtualmin")
	whm.ApiToken = getEnv("API_TOKEN", "")
	if whm.ApiToken == "" && serverType == "cpanel" {
		log("Please declare -x API_TOKEN=...")
		log("Other environments variables: MAX_PER_MIN=8 , MAX_PER_HOUR=100")
		log("NOTIFY_EMAIL=email , PF_LOG=/var/log/mail.log")
		log("WHM_API_HOST=127.0.0.1")
		log("SERVERTYPE=cpanel")
	}

	maxPerMin := int16(8)
	maxPerHour := int16(100)
	if os.Getenv("MAX_PER_MIN") != "" {
		if i, err := strconv.ParseInt(os.Getenv("MAX_PER_MIN"), 10, 16); err != nil {
			panic(fmt.Errorf("Failed parsing MAX_PER_MIN: %+v", err))
		} else {
			maxPerMin = int16(i)
		}
	}
	if os.Getenv("MAX_PER_HOUR") != "" {
		if i, err := strconv.ParseInt(os.Getenv("MAX_PER_HOUR"), 10, 16); err != nil {
			panic(fmt.Errorf("Failed parsing MAX_PER_HOUR: %+v", err))
		} else {
			maxPerHour = int16(i)
		}
	}

	if maxPerHour < maxPerMin {
		panic(fmt.Errorf("Max per hour must be above max per minutes"))
	}

	if os.Getenv("PF_LOG") != "" {
		logFile = os.Getenv("PF_LOG")
	}

	if os.Getenv("NOTIFY_EMAIL") != "" {
		notifyEmail = os.Getenv("NOTIFY_EMAIL")
	}

	if os.Getenv("WHM_API_HOST") != "" {
		whm.ApiHost = os.Getenv("WHM_API_HOST")
	}

	whm.Log = log
	virtualmin.Log = log

	if len(os.Args) < 2 {
		log("args: start|run|skip|reset|suspend|unsuspend|info|help|test-notify|rerun")
		return
	}

	maxRun := -1
	now := time.Now()
	skipLastLine := false
	//start from yesterday min
	startTime := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Local().Location())
	switch os.Args[1] {
	case "reset":
		log("Removing %s*", dataPath)
		tools.RemoveSubFileFolder(dataPath)
		os.Remove(configPath)
		log("Removed %s", configPath)
		return
	case "start":
		//use yesterday
	case "run":
		maxRun = 1
		startTime = now
	case "rerun":
		if len(os.Args) < 3 {
			log("rerun date/time, eg: 2022 oct 10")
			return
		}
		log("date? %v", os.Args)
		dateStr := fmt.Sprintf("%s %s %s", os.Args[2], os.Args[3], os.Args[4])
		thetime, err := tools.ParseDate(dateStr)
		if err != nil {
			panic(fmt.Errorf("Unable to read date: %#v", dateStr))
		}
		log("Rerun from: %s", thetime.Format(time.RFC3339))
		if err := cleanupFrom(thetime); err != nil {
			log(fmt.Sprintf("Unable to cleanup time: %+v", err))
		}

		startTime = thetime
		skipLastLine = true

	case "skip":
		startTime = time.Now() //skip to now, skip everything then...
	case "suspend":
		if len(os.Args) < 3 {
			log("suspend [email]")
			return
		}
		email := os.Args[2]
		if serverType == "cpanel" {
			if err := whm.SuspendEmail(email); err != nil {
				panic(fmt.Sprintf("error: %+v", err))
			}
		}
		if serverType == "virtualmin" {
			if err := virtualmin.SuspendEmail(email); err != nil {
				panic(fmt.Sprintf("error: %+v", err))
			}
		}

		log("Suspended %s", email)
		return
	case "unsuspend":
		if len(os.Args) < 3 {
			log("unsuspend [email]")
			return
		}

		email := os.Args[2]
		if err := virtualmin.EnableEmail(email); err != nil {
			panic(fmt.Sprintf("error: %+v", err))
		}
		log("Unsuspended %s", email)

		return
	case "info":
		if len(os.Args) < 3 {
			log("info [domain]")
			return
		}
		if serverType == "virtualmin" {
			panic(fmt.Errorf("Not implemented"))
		}
		info, err := whm.UserDataInfo(os.Args[2])
		if err != nil {
			panic(fmt.Sprintf("error: %+v", err))
		}
		log("%#v", info)
		return
	case "test-notify":
		if err := notifySuspend("test@example.com", "a test"); err != nil {
			log("notifySuspend error: %+v", err)
		}
		return

	case "help":
		log("start - continue from last position or start from yesterday, and repeats from last position")
		log("rerun - rerun from specified date")
		log("run - continue from last position or start from beginning for one time")
		log("skip - skip all existing data and repeats for new logs")
		log("reset - reset all data, huh, what?")
		log("suspend - suspend outgoing email")
		log("unsuspend - unsuspend outgoing email")
		log("info - get information of a domain")
		log("test-notify - test send notification mail")
		log("help - this!")
		return

	default:
		panic(fmt.Errorf("Unknown command: %s", os.Args[1]))
	}

	i := 1
	for {
		log("loop %d", i)
		if err := postfixLogScanner(logFile, startTime, maxPerMin, maxPerHour, skipLastLine); err != nil {
			log("log scanner error: %+v", err)
			// time.sleep(15 * time.Second)
		}

		if maxRun > -1 && i > maxRun {
			break
		}
		time.Sleep(15 * time.Second)
		i++
	} //loop

	log("Done.")
}

func cleanupFrom(thetime time.Time) error {
	t := thetime
	log("cleaningFrom: %v", t.Format(time.RFC3339))
	d, err := os.Open(dataPath)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		ownerDir := filepath.Join(dataPath, name)
		log("scanning: %+v", ownerDir)

		d, err := os.Open(ownerDir)
		if err != nil {
			return err
		}
		defer d.Close()
		dates, err := d.Readdirnames(-1)
		if err != nil {
			return err
		}
		for _, date := range dates {
			dateDir := filepath.Join(ownerDir, date)

			dirtime, err := time.Parse("2006_01_02", date)
			if err != nil {
				panic(fmt.Errorf("Unable to read date: %#v", date))
			}

			if !t.After(dirtime) {
				log("removing: %+v", dateDir)

				err = os.RemoveAll(dateDir)
				if err != nil {
					return err
				}
			} else {
				log("not removing: %+v", dateDir)
			}

		} //each date

	} //each owner

	// t = t.AddDate(0, 0, 1)
	return nil
}

func emailDomainName(email string) (string, error) {
	trimmedEmail := strings.TrimSpace(email)
	lastPos := strings.LastIndex(trimmedEmail, "@")
	if lastPos < 0 {
		return "", fmt.Errorf("@ missing, not valid email: " + email)
	}

	domain := trimmedEmail[lastPos+1:]
	log("domain %s, email: %s", domain, email)

	return domain, nil
}

func postfixLogScanner(logFile string, startTime time.Time, maxPerMin int16, maxPerHour int16, skipLastLine bool) error {
	lastLine := int64(0)
	lastPrefix := ""

	if !skipLastLine {
		var err error
		_, lastLine, lastPrefix, err = lastConfig(logFile)
		if err != nil {
			panic(err)
		}

		lastPrefix = strings.TrimRight(lastPrefix, "\n")
	}

	ignoreList := []string{}
	if (tools.FileExists("skip.conf")) {
		var err error
		ignoreList, err = tools.ReadContentAsList("skip.conf")
		if err != nil {
			panic(err)
		}
	}

	newSize := MustSize(logFile)
	file, err := os.Open(logFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	//checks if line still valid
	scanner := bufio.NewScanner(file)
	lineNo := int64(1)

	if lastPrefix != "" {
		var err error
		log("parsing last time: %s", lastPrefix[:19])
		startTime, err = tools.ParseDate(lastPrefix[:19])
		if err != nil {
			log("Unable to read lastPrefix date: %#v on line %d", startTime, lastPrefix)
			// panic(fmt.Errorf("Unable to read lastPrefix date: %#v on line %d", startTime, lastPrefix))
			startTime = time.Now()
		}
	}

	for scanner.Scan() {
		if lineNo >= lastLine {
			break
		}

		// log("Skipping: %d: %d", lineNo, scanner.Text())
		lineNo++
	}

	log("Scanning log from time: %v, last line %v path: %v", startTime.Format(time.RFC3339), lastLine, logFile)

	text := scanner.Text()
	if lastLine > 0 {
		if !strings.HasPrefix(text, lastPrefix) {
			log("Last missing line:\n%s\nExpecting:\n%s\n", text, lastPrefix)
			time.Sleep(10 * time.Second)
			//resetting to start
			_, lastLine, lastPrefix = 0, 0, ""
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return err
			}
			scanner = bufio.NewScanner(file) //reset scanner
			scanner.Scan()
			lineNo = 1
		} else {
			//skip to next line
			if !scanner.Scan() {
				log("No new line since last scan")
				return nil
			}
			lineNo++
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	log("Starting line %d time: %v", lineNo, startTime.Format(time.RFC3339))
	currentYear := fmt.Sprintf("%d ", time.Now().Year()) 
	currentTimeStr := ""
	currentSessionId := ""
	currentSender := ""
	currentRecipients := []string{}
	for {
		text = scanner.Text()
		log("raw line %d: %v", lineNo, text)
		res := postfixRegLine.FindStringSubmatch(text)
		// 0 = datetime, 1 = session-id, 2 = from/to, 3 = email
		debugList := res
		if (res != nil) {
			debugList = res[1:]
		}
		jsonStuff, _ := json.Marshal(debugList)
		log("regex: count %d %s", len(res), string(jsonStuff))
		// 0 - fullline
		// 1 - date
		// 2 - session id
		// 3 - from/to
		// 4 - email
		if len(res) < 5 {
			// if strings.Contains(text, "A=dovecot") {
			// log("Not: %#v | %v", res, text)
			// 	time.Sleep(100 * time.Millisecond)
			// }
		} else {
			if (res[3] == "from") {

				// execute last batch session
				if (currentSender != "" && len(currentRecipients) > 0) {
					log("Executing session: %s", currentSessionId)

					err := processSession(maxPerMin, maxPerHour, lineNo, currentSessionId, currentSender, currentRecipients, startTime, currentTimeStr, ignoreList)
					if (err != nil) {
						log("Error processing session %s: %s", currentSessionId, err)
					}
				}

				log("currentYear: %s", currentYear)
				
				currentTimeStr = currentYear + res[1]
				currentSender = res[4]
				currentSessionId = res[2]
				currentRecipients = []string{}

				currentDateTime, err := tools.ParseDate(currentTimeStr)
				if err != nil {
					log("Error parsing date: %s", currentTimeStr)
					return err
				}
				if currentDateTime.After(time.Now().Add(1 * time.Hour)) {
					currentTimeStr = fmt.Sprintf("%d %s", time.Now().Year(), res[1])
				}

			} else if (res[3] == "to") {
				// check session id, otherwise ignore
				if (currentSessionId != res[2]) {
					log("Session id mismatch: %s != %s", currentSessionId, res[2])
				} else {
					currentRecipients = append(currentRecipients, res[4])
				}
				
			} else {
				log("Not regex: %#v | %v", res, text)
			}
		}
		// log("read line %d", lineNo)
		if !scanner.Scan() {
			// log("scanned ended: line %d", lineNo)
			if (currentSender != "" && len(currentRecipients) > 0) {
				log("Executing final session: %s", currentSessionId)

				if err := processSession(maxPerMin, maxPerHour, lineNo, currentSessionId,
					currentSender, currentRecipients, startTime, currentTimeStr, ignoreList); err != nil {
					log("Error processing session %s: %s", currentSessionId, err)
				}
			}

			break
		}
		lineNo++
	}

	log("ended: line %d", lineNo)
	lastPrefix = text
	if (len(text) > 25) {
		lastPrefix = text[0:25]
	}

	storeConfig(logFile, newSize, lineNo, lastPrefix)
	return nil
}

func processSession(maxPerMin int16, maxPerHour int16, lineNo int64,
	sessionId string, sender string, recipients []string,
	startTime time.Time, timeStr string, ignoreList []string) error {
	log("Processing session %s: %s -> %v", sessionId, sender, recipients)
	
	var thetime time.Time
	var err error
	skipTime := false
	thetime, err = tools.ParseDate(timeStr)
	if err != nil {
		panic(fmt.Errorf("Unable to read date: %#v on line %d", thetime, lineNo))
	}
	if !startTime.IsZero() {
		if thetime.Before(startTime) {
			log("Skipping by time %s expected %s", thetime.Format(time.RFC3339), startTime.Format(time.RFC3339))
			skipTime = true
		} else {
			// log("ok time %s(%s) after %s", thetime.Format(time.RFC3339), res[1], startTime.Format(time.RFC3339))
		}
	} else {
		log("Start time is zero? %s", startTime.Format(time.RFC3339))
	}

	if (skipTime) {
		log("Skiptime session %s: %s -> %v", sessionId, sender, recipients)
		return nil
	}

	process := strings.Index(sender, "@") > 0
	if process {
		for _, ignore := range ignoreList {
			if strings.HasPrefix(ignore, "*") {
				if strings.HasSuffix(sender, ignore[1:]) {
					process = false
					break
				}
			}

			if strings.HasSuffix(ignore, "*") {
				if strings.HasSuffix(sender, ignore[:1]) {
					process = false
					break
				}
			}

			if !strings.HasPrefix(ignore, "*") && !strings.HasSuffix(ignore, "*") {
				if sender == ignore {
					process = false
					break
				}
			}
		}
	}
	
	if (!process && !skipTime) {
		log("Ignoring session %s %#v : %#v", sessionId, sender, thetime.Format(time.RFC3339))
		time.Sleep(2 * time.Second)
		return nil
	}

	senderDomain, err := emailDomainName(sender)
	if err != nil {
		return fmt.Errorf("unable to obtain domain from email %s, error: %v", sender, err.Error())
	}
	hasExternal := false
	
	r := 0
	for {
		if r >= len(recipients) {
			break
		}
		recipient := recipients[r]
		log("session %s: %s -> %s", sessionId, sender, recipient)
		
		recipientDomain, err := emailDomainName(recipient)

		if err != nil {
			log(fmt.Sprintf("unable to obtain domain from email %s, error: %v", err, recipient))
			// return fmt.Errorf("unable to obtain domain from email %s, error: %v", err, recipient)
			r = r + 1
			continue
		}
		if senderDomain == recipientDomain {
			log("detected same domain %s | %s", sender, recipient)
			r = r + 1
			continue
		}
		log("detected other domain %s | %s", recipientDomain, recipient)
		hasExternal = true
		r = r + 1
	}

	if hasExternal {
		log("detected external recipient of %s", sender)
		// time.Sleep(2 * time.Second)
		minCount, hourCount, err := mailCount(thetime, sender)
		if err != nil {
			return err
		}
		
		minCount++
		hourCount++
		//TODO save based on X recipients per email?

		if err := mailCountStore(thetime, sender, hourCount, minCount); err != nil {
			panic(fmt.Errorf("Unable to save count %s, time: %#v, error: %#v", sender, thetime, err))
		}

		if minCount > int64(maxPerMin) || hourCount > int64(maxPerHour) {
			if err := virtualmin.SuspendEmail(sender); err != nil {
				log("Unable to suspendEmail %s, error: %+v", sender, err)
				time.Sleep(5 * time.Second)
			}

			if notifyEmail != "" {
				if err = notifySuspend(sender, fmt.Sprintf("Count: minute: %d, hour: %d", minCount, hourCount)); err != nil {
					log("notifySuspend error: %+v", err)
					time.Sleep(10 * time.Second)
				}
			}
		}

		log("Written %s time: %v, min: %v, hour: %v", sender, thetime, minCount, hourCount)
	}
	
	log("Done session %s: %s -> %v", sessionId, sender, recipients)
	time.Sleep(1 * time.Second)
	return nil
}

func notifySuspend(email string, message string) error {
	if notifyEmail == "" {
		return fmt.Errorf("NOTIFY_EMAIL not set")
	}

	c1 := exec.Command("echo", "-e", fmt.Sprintf("\"%s\"", message))
	c2 := exec.Command("mail", "-s", fmt.Sprintf("\"suspended email %s\"", email), notifyEmail)
	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r
	var b2 bytes.Buffer
	c2.Stdout = &b2

	if err := c1.Start(); err != nil {
		return err
	}

	if err := c2.Start(); err != nil {
		return err
	}

	if err := c1.Wait(); err != nil {
		return err
	}

	w.Close()

	if err := c2.Wait(); err != nil {
		return err
	}

	io.Copy(os.Stdout, &b2)

	log("mail-result: %s", b2.Bytes())
	return nil
}

func mailCountStore(thetime time.Time, email string, hourCount int64, minCount int64) error {
	path := dataPath + cleanPath(email)

	dirPath := cleanPath(thetime.Format("2006-01-02"))
	hourPath := thetime.Format("15")
	minPath := thetime.Format("1504")

	datePath := path + "/" + dirPath
	hourFile := datePath + "/" + hourPath
	minFile := datePath + "/" + minPath

	MustDir(datePath)

	log("Writing %s/%s", dirPath, hourFile)
	if err := ioutil.WriteFile(hourFile, []byte(fmt.Sprintf("%d", hourCount)), 0644); err != nil {
		return err
	}
	log("Writing %s", minFile)
	if err := ioutil.WriteFile(minFile, []byte(fmt.Sprintf("%d", minCount)), 0644); err != nil {
		return err
	}
	return nil
}

// this minute, this hour count
func mailCount(thetime time.Time, email string) (int64, int64, error) {
	path := dataPath + cleanPath(email)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/whatever does not exist
		return 0, 0, nil
	}

	dirPath := cleanPath(thetime.Format("2006-01-02"))
	// hourPath := now.Format("150405")
	hourPath := thetime.Format("15")
	minPath := thetime.Format("1504")

	datePath := path + "/" + dirPath
	if _, err := os.Stat(datePath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		return 0, 0, nil
	}

	if _, err := os.Stat(datePath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		return 0, 0, nil
	}

	hourCount := int64(0)
	minCount := int64(0)
	hourFile := datePath + "/" + hourPath
	minFile := datePath + "/" + minPath

	if _, err := os.Stat(hourFile); !os.IsNotExist(err) {
		content, err := ioutil.ReadFile(hourFile)
		if err != nil {
			return 0, 0, err
		}
		hourCount, err = strconv.ParseInt(string(content), 0, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	if _, err := os.Stat(minFile); !os.IsNotExist(err) {
		content, err := ioutil.ReadFile(minFile)
		if err != nil {
			return 0, 0, err
		}
		minCount, err = strconv.ParseInt(string(content), 0, 64)
		if err != nil {
			return 0, 0, err
		}
	}

	return minCount, hourCount, nil
}

func storeConfig(logFile string, size int64, line int64, prefix string) error {
	return ioutil.WriteFile(configPath, []byte(fmt.Sprintf("%d||%d||%s", size, line, prefix)), 0644)
}

//last size, last line#, prefix
func lastConfig(logFile string) (int64, int64, string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		return 0, 0, "", nil
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return 0, 0, "", err
	}

	args := strings.Split(string(content), "||")
	size, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return 0, 0, "", err
	}
	line, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return 0, 0, "", err
	}
	prefix := args[2]

	newSize := MustSize(logFile)
	if newSize < size {
		log("Size shrinked: %f, expected: %f", newSize, size)
		time.Sleep(10 * time.Second)
		return 0, 0, "", nil //reset
	}

	return size, line, prefix, nil
}

func cleanPath(name string) string {
	res := strings.Replace(name, "@", "_", -1)
	res = strings.Replace(res, "-", "_", -1)
	res = filepath.Clean(res)
	return res
}

func MustDir(path string) {
	log("MustDir: %s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log("Creating dir: %s", path)
		if err := os.MkdirAll(path, 0744); err != nil {
			panic(fmt.Errorf("Unable to create %s, error: %+v", path, err))
		}
	} else if err != nil {
		panic(fmt.Errorf("Unknown mustdir %s, error: %+v", path, err))
	}
}

func MustSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		panic(fmt.Errorf("Unable to state %s, error: %+v", path, err))
	}

	return fi.Size()
}

func log(msg string, args ...interface{}) {
	fmt.Printf("pfmon(v1.0.0):"+msg+"\n", args...)
}
