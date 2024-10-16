package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

/** Client initialization. */
var client *Client

func initClient() error {
	var err error
	client, err = New()
	return err
}

/** Retrieve all the notes. */
var (
	orderBy    string
	orderDir   string
	getAllNote = &cobra.Command{
		Use:     "getallnotes [field1,...,fieldn]",
		Short:   "Retrieve all the notes. Optionally specify which fields and in which order.",
		Aliases: []string{"getall"},
		Args:    cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			note, err := client.GetAllNotes(strings.Join(args, ","), orderBy, orderDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			data, _ := json.MarshalIndent(note, "", "")
			fmt.Println(string(data))
		},
	}
)

/** Retrieve a single note. */
var getNote = &cobra.Command{
	Use:     "getnote note_id [field1,...,fieldn]",
	Short:   "Retrieve a note with a given ID. Optionally specify which fields to return.",
	Aliases: []string{"get"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		note, err := client.GetNote(args[0], strings.Join(args[1:], ","))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		data, _ := json.MarshalIndent(note, "", "")
		fmt.Println(string(data))
	},
}

func Execute() {
	if err := initClient(); err != nil {
		fmt.Println("Error initializing client:", err)
		os.Exit(1)
	}

	getAllNote.Flags().StringVarP(&orderBy, "order_by", "p", "", "order by field")
	getAllNote.Flags().StringVarP(&orderDir, "order_dir", "d", "", "order direction (asc/desc)")

	rootCmd := &cobra.Command{Use: "joplingo-cli"}
	rootCmd.AddCommand(getNote, getAllNote)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
