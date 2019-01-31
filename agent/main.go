package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

	cmds := make(map[int]*exec.Cmd)

	fmt.Printf("device-id=%s\n", deviceID)

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", mqttServer))
	opts.SetUsername("79412d37892ec50b79d2c17fcb311c5db9dc262429b46f8d3925fa20c2392533")
	opts.SetPassword("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIvMFx1MDAxZFx1MDAxOVx1ZmZmZFx1MDAxZXNvXHUwMDE3XHUwMDFlXHUwMDFiXHUwMDA1XHVmZmZkO0hcdWZmZmRcdTAwMGJcdTAwMTFcdWZmZmRvXHUwMDExXHVmZmZkXHVmZmZkXHVmZmZkXHVmZmZkNVx1ZmZmZHvMgtWbIiwiaWF0IjoxNTQ4NzgyNzY2LCJpc3MiOiJMZXRBdXRoIiwibmJmIjoxNTQ4NzgyNzY2LCJzdWIiOiJBbGwifQ.ABuZqjHA8aIifFWoKoFrBEQOW_67MR0SxrWNsJWq0H0XFAqfOuCq_9fou1H57p3hubHftl-M_N59obQI45YqeJBn8XYUWJjM780t7FeRkJgGcJ935i7mhXUetCG2gdwvPosfgHfVLb_RkfRlKI3LnqkO9YwlnzmDrKV9NmoG_ZV_K-U-6GQER9cAnit-dVKXV-rBrWXs1XiUXnhYoLFWjmlXBz48SUqfUrBjx04L-DRN3Te3rsZlEm1pGUjGL-tQJVDJmZHmvlnLTPFtXnGxMIMAG3uk4XLT4MUyKg2YMrF6h6mbTvBD_9onwPukx7db8DfWgwmdmKuWIwpOvplqHQDSeGhuGPHJcUvTdqdiNTOKvYZz2rZyGiG6LGcAvbi5BiRXzUzNtK1-hfgRYBVDY7HZv3qWpzTl2U1S77JAzk1A6Yhi6CGAVKrswv9-pLECagPuw7S2h8dt6c2hzkgnmJPXn48LsujptMePhmjWLrahq-A_Mq6k8WEGA-hkhgdnnlcOqqlKgFe5CHZScJSd4MXwkctHqa519wRapEbt6bFMqcVFVccTmr_j5c0jtf-QRo2Vktg_angTCaH6NZYb_YM8K8h6atj-g5JuJ3QGlCBn8_9Mk_GDcDKZS2CobBbUk40LQFrMuZmSzJ9mGGe6L5cPHYUNXg60iRjEZkaXHv8")
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
