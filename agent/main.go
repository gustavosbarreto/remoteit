package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/journal"
	"github.com/coreos/go-systemd/sdjournal"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type JournalState struct {
	Timestamp time.Time
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "ssh-tunnel",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	viper.AutomaticEnv()

	deviceID := getDeviceID(viper.GetString("DEVICE_ID"))
	mqttServer := viper.GetString("MQTT_SERVER")
	sshServer := viper.GetString("SSH_SERVER")
	sshPort := viper.GetString("SSH_PORT")
	privateKey := viper.GetString("PRIVATE_KEY")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	helpCalled, err := rootCmd.Flags().GetBool("help")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if helpCalled {
		os.Exit(1)
	}

	j, err := sdjournal.NewJournal()

	state := &JournalState{}

	b, err := ioutil.ReadFile("state.dat")
	if err == nil {
		r := bytes.NewReader(b)
		dec := gob.NewDecoder(r)
		dec.Decode(state)
	}

	if !state.Timestamp.IsZero() {
		j.SeekRealtimeUsec(uint64(state.Timestamp.UnixNano() / 1000))
	}

	go func() {
		for {
			n, err := j.Next()
			if err != nil && err != io.EOF {
				panic(err)
			}

			if n < 1 {
				// no new entry
				j.Wait(sdjournal.IndefiniteWait)
				continue
			}

			entry, err := j.GetEntry()
			if err != nil {
				panic(err)
			}

			var l struct {
				Message   string    `json:"message"`
				Timestamp time.Time `json:"timestamp"`
				Level     string    `json:"level"`
			}

			l.Message = entry.Fields[sdjournal.SD_JOURNAL_FIELD_MESSAGE]
			l.Timestamp = time.Unix(0, int64(entry.RealtimeTimestamp*1000))

			level, err := strconv.Atoi(entry.Fields[sdjournal.SD_JOURNAL_FIELD_PRIORITY])
			if err != nil {
				panic(err)
			}

			levels := map[journal.Priority]string{
				journal.PriEmerg:   "emerg",
				journal.PriAlert:   "alert",
				journal.PriCrit:    "crit",
				journal.PriErr:     "err",
				journal.PriWarning: "warning",
				journal.PriNotice:  "notice",
				journal.PriInfo:    "info",
				journal.PriDebug:   "debug",
			}

			l.Level = levels[journal.Priority(level)]

			request := gorequest.New()
			_, _, errs := request.Post("http://localhost/log/log").Send(l).End()
			if len(errs) > 0 {
				j.Previous()
				continue
			}

			state.Timestamp = l.Timestamp

			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)

			enc.Encode(state)

			ioutil.WriteFile("state.dat", buf.Bytes(), 0644)

			time.Sleep(time.Millisecond * 100)
		}
	}()

	cmds := make(map[int]*exec.Cmd)

	fmt.Printf("device-id=%s\n", deviceID)

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", mqttServer))
	opts.SetUsername(deviceID)
	opts.SetPassword("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJcdWZmZmR8XHVmZmZkXHVmZmZkXHUwMDBlMn9cdWZmZmRcdWZmZmRcdWZmZmRcdWZmZmRcdWZmZmQuXHUwMDEza1x1ZmZmZFx1ZmZmZFx1ZmZmZFxuXHVmZmZkXHUwMDNjXHVmZmZkXHVmZmZkRFx1ZmZmZFx1ZmZmZDhcdWZmZmTUkFx1MDAwMzYifQ.npbhjNu0b8gkFYx59GxAyDDN2nSAq6oS1Qxt3Quvb3cfLaWVpyE0RxkkAOo7lvwgb6rEQC4qjpmQx-K95ZOM76jC1O3v3370MNn5lgV4Dk5UGlJ5Yotfhc2LQIujotFcA1ha1KyvCVaOf2s-y0OkK0O188XeXsS4D5zFILR7JnbiIRkdVh1TBDLGA2JHRiemIN4QXPvxeybZFARNZaRKTcrdaqr5VUbZn7lh_VOyfP9NdMQzBEsF2AQuIcYbE9JebmBO_TEdZWLiMexnorT6iGW506qk8dhpBS_LQYszGulVgfwQ7s18Yh_WdZYCzwErfAf2JA330eiOY7p6rH02cUGW6fBpbUG-TuAGMftlKQTrpyS1yQtEhUNKiJ8hKkEOkEy1tdmF5-CfDOWJKlpTg0wzhZSraXSiXPZA33BoyC2ZIEORE8GN5yUO2vCY8DiqNHVBlaK8e1Ti7DCdjm-Es8PVGjfpTHTmQFlbatEuLy2sLj3t7MrO_Drofcn9sBp9iT-l1Pm-Zj_VNunjp_MnoNfet0MfOwz7CPZ3tFRWPZCCTtHSj2zEBm2fmYOfFWuc6o9EYuHtKGD-v_0W5cw3kmWpMZYalhMHbOrAcLwQYiBHMdW4UoZNQSKKHfGk80EIHvWzwM9Xaf_HRDk3XZVK6uYf0yHA1gv_9SZ_SQTZFuo")
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		connect(client)
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if token := client.Subscribe(fmt.Sprintf("connect/%s", deviceID), 0, func(client mqtt.Client, msg mqtt.Message) {
			go func() {
				parts := strings.SplitN(string(msg.Payload()), ":", 2)
				port, _ := strconv.Atoi(parts[0])

				fmt.Printf("reverse port=%d\n", port)

				args := []string{
					"ssh",
					"-i", privateKey,
					"-o", "StrictHostKeyChecking=no",
					"-nNT",
					"-p", sshPort,
					"-R", fmt.Sprintf("%d:localhost:22", port),
					fmt.Sprintf("%d:%s@%s", port, parts[1], sshServer),
				}

				cmd := exec.Command(args[0], args[1:]...)
				_ = cmd.Start()

				cmds[port] = cmd
			}()
		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}

		if token := client.Subscribe(fmt.Sprintf("disconnect/%s", deviceID), 0, func(client mqtt.Client, msg mqtt.Message) {
			port, _ := strconv.Atoi(string(msg.Payload()))

			if cmd, ok := cmds[port]; ok {
				cmd.Process.Kill()
				cmd.Wait()
				delete(cmds, port)
			}
		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}
	})

	client := mqtt.NewClient(opts)

	connect(client)

	select {}
}

func getDeviceID(deviceID string) string {
	parts := strings.Split(deviceID, ":")
	if len(parts) < 2 {
		return deviceID
	}

	switch parts[0] {
	case "value":
		return strings.Join(parts[1:], ":")
	case "exec":
		args, err := shellwords.Parse(strings.Join(parts[1:], ":"))
		if err != nil {
			log.Fatal(err)
		}

		var out bytes.Buffer

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		return strings.TrimSpace(out.String())
	}

	return deviceID
}

func connect(client mqtt.Client) {
	for {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			time.Sleep(time.Second)
			continue
		}

		break
	}
}
