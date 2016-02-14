// Copyright © 2016 Luit van Drongelen <luit@luit.eu>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd // import "luit.eu/rcp/cmd"

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"luit.eu/rcp/server"
)

var cfgFile string

// rootCmd is the `rcp` command
var rootCmd = &cobra.Command{
	Use:   "rcp",
	Short: "Redis Cluster Proxy for cluster-unaware software",
	Long: `Redis Cluster Proxy is a daemon to help your application to work with Redis
Cluster without cluster-aware code. This can be useful if you can't or won't
change the application's code. All you have to do is make sure you don't
issue commands that are impossible (commands accessing across hash slots).`,
	Run: func(cmd *cobra.Command, args []string) {
		addrstr := fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port"))
		laddr, err := net.ResolveTCPAddr("tcp", addrstr)
		if err != nil {
			fmt.Printf("Error: unable to use address %s as TCP address: %v", addrstr, err)
			return
		}
		l, err := net.ListenTCP("tcp", laddr)
		if err != nil {
			fmt.Printf("Error: unable to listen on %s: %v", laddr.String(), err)
			return
		}
		defer l.Close()
		fmt.Printf("Listening on %v\n", laddr)
		for {
			c, err := l.AcceptTCP()
			if err != nil {
				fmt.Printf("Error: accept: %v", err)
				return
			}
			go server.Dumb(c)
		}
	},
}

// Execute activates the `rcp` command. This is called by main.main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(64)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.rcp.yaml)")

	rootCmd.PersistentFlags().IPP("bind", "b", net.IPv4(127, 0, 0, 1), "IP address to bind to")
	viper.BindPFlag("bind", rootCmd.PersistentFlags().Lookup("bind"))

	rootCmd.PersistentFlags().IntP("port", "p", 6379, "Port to listen on")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".rcp")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("rcp")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.UnsupportedConfigError); ok {
			// Probably no config found
		} else {
			fmt.Printf("Unable to read config: %v\n", err)
		}
	}
}
