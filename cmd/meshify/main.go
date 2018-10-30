package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tniswong/meshify/pkg/etl"
	"github.com/tniswong/meshify/pkg/twitter"
	"log"
	"os"
	"strings"
)

// MeshifyConfig stores the values that may be passed in via command line or environment variables
type MeshifyConfig struct {
	Key      string
	Secret   string
	Out      *os.File
	Hashtags []string
	N        int
}

// RootCommand is the root cobra command
var RootCommand = &cobra.Command{
	Use:   "meshify",
	Short: "Todd Niswonger's technical evaluation exercise for Meshify",
	Long:  `Use the Twitter API to gather 2000 unique tweets with the hashtag #IoT and output them to a CSV file.`,
	Run: func(cmd *cobra.Command, args []string) {

		c, err := configure()

		if err != nil {
			log.Fatal(err)
		}

		defer c.Out.Close()

		api := twitter.NewAPI(c.Key, c.Secret)
		e := etl.HashtagsToCSV(c.Out, api, c.N, c.Hashtags...)

		if err := e.ETL(); err != nil {
			log.Fatal(err)
		}

	},
}

func init() {

	log.SetFlags(0)

	// register flags to command
	RootCommand.PersistentFlags().StringP("api-key", "k", "", "Required. Twitter API Public Key. If unset uses MESHIFY_API_KEY environment variable.")
	RootCommand.PersistentFlags().StringP("api-secret", "s", "", "Required. Twitter API Secret Key. If unset uses MESHIFY_API_SECRET environment variable.")
	RootCommand.PersistentFlags().StringP("out", "o", "", "Output file path for csv formatted output. (default STDOUT)")
	RootCommand.PersistentFlags().StringSliceP("tags", "t", []string{"IoT"}, "Hashtags to query. Note: '#' is a comment character in bash, so use ONLY the tag name rather than the full hashtag (ex: 'IoT' NOT '#IoT'!)")
	RootCommand.PersistentFlags().IntP("number", "n", 2000, "Number of tweets per hashtag.")

	// required flags
	RootCommand.MarkFlagRequired("api-key")
	RootCommand.MarkFlagRequired("api-secret")

	// bind flags to viper
	viper.BindPFlags(RootCommand.PersistentFlags())

	// environment variable fallbacks
	viper.SetEnvPrefix("MESHIFY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

}

func main() {

	if err := RootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func configure() (MeshifyConfig, error) {

	k := viper.GetString("api-key")
	if k == "" {
		return MeshifyConfig{}, errors.New("error: flag [-k, --api-key string] or environment variable MESHIFY_API_KEY is required")
	}

	s := viper.GetString("api-secret")
	if s == "" {
		return MeshifyConfig{}, errors.New("error: flag [-s, --api-secret string] or environment variable MESHIFY_API_SECRET is required")
	}

	var hashtags []string
	tags := viper.GetStringSlice("tags")
	for _, hashtag := range tags {
		hashtags = append(hashtags, "#"+hashtag)
	}

	c := MeshifyConfig{
		Key:      k,
		Secret:   s,
		Out:      os.Stdout,
		Hashtags: hashtags,
		N:        viper.GetInt("number"),
	}

	o := viper.GetString("out")

	if o != "" {

		outfile, err := os.Create(o)

		if err != nil {
			return MeshifyConfig{}, fmt.Errorf("error: could not open file: %v", err)
		}

		c.Out = outfile

	}

	return c, nil

}
