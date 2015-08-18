package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/term1nal/smallcfg"
)

type Config struct {
	SkipList []string `json:"skiplist"`
}

var config Config

var configFile = "'/etc/vcp-spam.conf.json"

func main() {
	log.Println("Loading configuration...")
	err := smallcfg.LoadConfig(configFile, &config)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Configuration Loaded!")

	clean := flag.Bool("clean", false, "Enables removal of the contents of the learn folders.")
	flag.Parse()

	log.Println("Initiating SpamAsssassin Bayesian Learning for Vesta Control Panel\n============================")
	log.Println("Running SpamAssassin update...")
	err = runCommand("/usr/bin/sa-update")
	checkErr(err, "SpamAssassin update command exited with errors (check Exit Codes documentation)", false)
	log.Println("SpamAssassin update process completed")

	log.Println("Parsing Mailboxes...")
	mboxes := getMailboxes()

	log.Println("Got Mailboxes, sending to SpamAssassin learning commands...")

	for _, mbox := range mboxes {
		junkpath := filepath.Join(mbox, ".Junk")
		hampath := filepath.Join(mbox, ".NotSpam")

		file, err := os.Stat(junkpath)

		var cmd string

		if checkErr(err, fmt.Sprintf("User @ %q does not have a \".Junk\" folder, skipping.", mbox), false) {
			goto skipjunk
		}

		if !file.IsDir() {
			log.Printf("Warning: %q is not a directory, skipping.", mbox)
			goto skipjunk
		}

		log.Printf("Learning junk mail from mailbox: %q", mbox)
		cmd = fmt.Sprintf("/usr/bin/sa-learn --spam %s/{new,cur}", hampath)
		err = runCommand(cmd)
		checkErr(err, "Unable to run SpamAssassin Learn command (likely empty folder)", false)

		if *clean {
			log.Printf("Cleaning messages from %q because 'clean' flag was set...", junkpath)
			curpath := filepath.Join(junkpath, "/cur")
			newpath := filepath.Join(junkpath, "/new")
			delcur := fmt.Sprintf("rm -f %s/*", curpath)
			delnew := fmt.Sprintf("rm -f %s/*", newpath)

			err = runCommand(delcur)
			checkErr(err, fmt.Sprintf("Unable to delete junk from %q", curpath), false)

			err = runCommand(delnew)
			checkErr(err, fmt.Sprintf("Unable to delete junk from %q", newpath), false)
		}

	skipjunk:

		file, err = os.Stat(hampath)

		if checkErr(err, fmt.Sprintf("User @ %q does not have a \".NotSpam\" folder, skipping.", mbox), false) {
			continue
		}

		if !file.IsDir() {
			log.Printf("Warning: %q is not a directory, skipping.", mbox)
			continue
		}

		log.Printf("Learning non-junk mail from mailbox: %q", mbox)
		cmd = fmt.Sprintf("/usr/bin/sa-learn --ham %s/{new,cur}", hampath)
		err = runCommand(cmd)
		checkErr(err, "Unable to run SpamAssassin Learn command (likely empty folder)", false)

		if *clean {
			log.Printf("Cleaning messages from %q because 'clean' flag was set...", hampath)
			curpath := filepath.Join(hampath, "/cur")
			newpath := filepath.Join(hampath, "/new")
			delcur := fmt.Sprintf("rm -f %s/*", curpath)
			delnew := fmt.Sprintf("rm -f %s/*", newpath)

			err = runCommand(delcur)
			checkErr(err, fmt.Sprintf("Unable to delete messages from %q", curpath), false)

			err = runCommand(delnew)
			checkErr(err, fmt.Sprintf("Unable to delete messages from %q", newpath), false)
		}
	}

	log.Println("Running SpamAssassin database sync...")
	err = runCommand("/usr/bin/sa-learn --sync")
	checkErr(err, "Unable to run SpamAssassin Sync command", true)

	log.Println("Restarting SpamAssassin service...")
	err = runCommand("/etc/init.d/spamassassin restart")
	checkErr(err, "Unable to run SpamAssassin service restart command", true)
	log.Println("Done!\n============================")
}

func getMailboxes() []string {
	mailboxes := []string{}

	users, err := ioutil.ReadDir("/home")
	checkErr(err, "Unable to get user directory listing", true)

	for _, user := range users {
		if skipUser(user.Name()) {
			continue
		}
		domainpath := filepath.Join("/home", user.Name(), "mail")

		log.Printf("Found user: %q", user.Name())

		domains, err := ioutil.ReadDir(domainpath)
		if checkErr(err, "Unable to get user mail domain directory listing (no mail directory/not vesta account?)", false) {
			continue
		}

		for _, domain := range domains {
			mboxespath := filepath.Join(domainpath, domain.Name())

			log.Printf("Found mail domain: %q", mboxespath)

			mboxes, err := ioutil.ReadDir(mboxespath)
			if checkErr(err, "Unable to get domain mailbox directory listing", false) {
				continue
			}

			for _, mbox := range mboxes {
				mboxpath := filepath.Join(mboxespath, mbox.Name())

				log.Printf("Found mailbox: %q", mboxpath)

				if checkErr(err, "Unable to get mailbox directory listing", false) {
					continue
				}

				mailboxes = append(mailboxes, mboxpath)
			}
		}
	}

	if len(mailboxes) < 1 {
		log.Fatalln("Error: No mailboxes found.")
	}

	return mailboxes
}

func skipUser(user string) bool {
	for _, u := range config.SkipList {
		if u == user {
			return true
		}
	}

	return false
}

func checkErr(err error, msg string, fatal bool) bool {
	if err != nil {
		if fatal {
			log.Fatalf("Error: %s: %s", msg, err)
		} else {
			log.Printf("Warning: %s: %s", msg, err)
			return true
		}
	}
	return false
}

func runCommand(cmd string) error {
	return exec.Command("/bin/sh", "-c", cmd).Run()
}
