package postfix

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

func main2() {
	// Get the email address from the command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Please provide an email address to search for")
		return
	}
	sender := os.Args[1]

	// Set the path to your Postfix log file
	file, err := os.Open("/var/log/mail.log")
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	// Set up the regular expression to match the sender's email address in the log
	re := regexp.MustCompile(fmt.Sprintf(`from=<%s>`, regexp.QuoteMeta(sender)))

	// Start reading the log file line by line in reverse order
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line matches the regular expression
		if re.MatchString(line) {
			// Prepend the line to the slice of matching lines
			lines = append([]string{line}, lines...)
		}
	}

	// Print out the matching lines in reverse order to show the latest logs first
	for _, line := range lines {
		// Extract the timestamp from the log line
		fields := strings.Fields(line)
		timestampStr := fields[0] + " " + fields[1]
		timestamp, err := time.Parse("Jan 2 15:04:05", timestampStr)
		if err != nil {
			log.Fatal("Error parsing timestamp:", err)
		}

		// Get the human-readable time ago prefix
		elapsed := time.Since(timestamp)
		prefix := humanTimeAgoPrefix2(elapsed)

		// Print the log line with the time ago prefix
		fmt.Printf("%s %s\n", prefix, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading file:", err)
	}
}

// humanTimeAgoPrefix returns a human-readable prefix for the elapsed time.
// It returns "just now" if the elapsed time is less than a minute, "X minutes ago"
// if the elapsed time is less than an hour, or "X hours ago" if the elapsed time
// is greater than or equal to an hour.
func humanTimeAgoPrefix2(elapsed time.Duration) string {
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
