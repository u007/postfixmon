package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	// Get the email address from the command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Please provide an email address to search for")
		return
	}
	sender := os.Args[1]
	fmt.Println("Finding:", sender)

	// Set the path to your Postfix log file
	// file, err := os.Open("/var/log/mail.log")
	file, err := os.Open("mail.log")
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	// Get the size of the file
	// fileInfo, err := file.Stat()
	// if err != nil {
	// 	log.Fatal("stat", err)
	// }
	// Get the file size to set the offset for reading the log file in reverse order
	// fileSize := fileInfo.Size()

	// Create a new reader to read the file backwards
	reader := bufio.NewReader(file)

	// Seek to the end of the file
	// _, err = file.Seek(fileSize, 0)
	// if err != nil {
	// 	log.Fatal("seek error", err)
	// }
	// goes to last charactor
	_, err = file.Seek(-2, io.SeekEnd)
	if err != nil {
		// If we've reached the beginning of the file, break out of the loop
		log.Fatal("start seek failed", err)
	}
	fmt.Println("Reading line...")
	// Set up the regular expression to match the sender's email address in the log
	reSender := regexp.MustCompile(fmt.Sprintf(`from=<%s>`, regexp.QuoteMeta(sender)))
	
	// Read the file backwards line by line
	var buf []byte
	for {
		// Seek to the previous newline character
		_, err := file.Seek(-2, io.SeekCurrent)
		if err != nil {
			// If we've reached the beginning of the file, break out of the loop
			if err == os.ErrInvalid {
				fmt.Println("End of file reached")
				break
			}
			log.Fatal("seek failed:", err)
		}

		// Read the next line backwards
		ch, err := reader.ReadByte()
		if err != nil {
			log.Fatal("readbyte failed", err)
		}
		fmt.Printf(string(ch))
		if ch == '\n' {
			line := string(buf)
			// fmt.Printf("found line: %s", line)
			if reSender.MatchString(line) {
				// Extract the timestamp from the log line
				fields := strings.Fields(line)
				timestampStr := fields[0] + " " + fields[1] + " " + fields[2]
				timestamp, err := time.Parse("Jan 2 15:04:05", timestampStr)
				if err != nil {
					log.Fatal("Error parsing timestamp:", err, timestampStr)
				}

				// Get the human-readable time ago prefix
				fmt.Println("timestamp: ", timestamp)
				elapsed := time.Since(timestamp)
				// is future, is likely last year
				if (elapsed < 0) {
					timestamp, err = time.Parse("Jan 2 15:04:05", timestampStr)
					break
				}
				prefix := humanTimeAgoPrefix(elapsed)

				// Highlight the sender's email address in the log line
				line = reSender.ReplaceAllString(line, fmt.Sprintf("\033[1m\033[31mfrom=<%s>\033[0m", sender))

				// Highlight the recipient's email address in the log line
				reRecipient := regexp.MustCompile(`to=<([^>]*)>`)
				line = reRecipient.ReplaceAllString(line, "\033[1mto=<$1>\033[0m")

				// Print the log line with the time ago prefix
				fmt.Printf("%s %s\n", prefix, line)
			}
			
			buf = nil
		}
		buf = append(buf, ch)
	}

	log.Println("Done.")
}


// humanTimeAgoPrefix returns a human-readable prefix for the elapsed time.
// It returns "just now" if the elapsed time is less than a minute, "X minutes ago"
// if the elapsed time is less than an hour, or "X hours ago" if the elapsed time
// is greater than or equal to an hour.
func humanTimeAgoPrefix(elapsed time.Duration) string {
	if elapsed < time.Minute {
		return "just now"
	} else if elapsed < time.Hour {
		minutes := int(elapsed.Minutes())
		return fmt.Sprintf("%d minutes ago", minutes)
	} else {
		hours := int(elapsed.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	}
}
