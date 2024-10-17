package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	client  *Client
	verbose bool
)

/** Client initialization. */
func initClient() error {
	var err error
	client, err = New()
	return err
}

/** Delete a note with a given ID. */
var (
	permanent  bool
	deleteNote = &cobra.Command{
		Use:     "deletenote id",
		Short:   "Delete a note.",
		Aliases: []string{"del"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := client.DeleteNote(args[0], permanent)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if verbose {
				fmt.Printf("Removed note with ID: %s\n", id)
			}
		},
	}
)

/** Update an existing note with a given ID. */
var updateNote = &cobra.Command{
	Use:     "updatenote note_id field1 value1 [... fieldn valuen]",
	Short:   "Update a note with a given ID. Specify which fields to update with the corresponding values.",
	Aliases: []string{"update"},
	Args:    cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		note, err := client.UpdateNote(args[0], args[1:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Print the updated note
		if verbose {
			result, _ := json.MarshalIndent(note, "", "")
			fmt.Println(string(result))
		}
	},
}

/** Create a new note from a local file. */
var (
	format     string
	remove     bool
	createNote = &cobra.Command{
		Use:     "createnote path_to_file",
		Short:   "Create a new note from an existing file. Optionally specify the format.",
		Aliases: []string{"new"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Read body from file
			body, err := os.ReadFile(string(args[0]))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Remove path and extension for title
			fileName := filepath.Base(args[0])
			note, err := client.CreateNote(strings.TrimSuffix(fileName, filepath.Ext(fileName)), format, string(body))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Print the new note
			if verbose {
				data, _ := json.MarshalIndent(note, "", "")
				fmt.Println(string(data))
			}

			if remove {
				if err := os.Remove(args[0]); err != nil {
					fmt.Println(err)
				}
			}
		},
	}
)

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

/** Retrieve a single note with a given ID. */
var getNote = &cobra.Command{
	Use:     "getnote note_id [field1 field2 ... fieldn]",
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

	rootCmd := &cobra.Command{Use: "joplingo-cli"}
	rootCmd.AddCommand(deleteNote)
	rootCmd.AddCommand(updateNote, createNote)
	rootCmd.AddCommand(getNote, getAllNote)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	deleteNote.Flags().BoolVarP(&permanent, "permanent", "p", false, "permanently delete the note")
	createNote.Flags().BoolVarP(&remove, "delete", "d", false, "delete the source file afterward")
	createNote.Flags().StringVarP(&format, "format", "f", "markdown", "format of the note (markdown/html)")
	getAllNote.Flags().StringVarP(&orderBy, "order_by", "f", "", "order by field")
	getAllNote.Flags().StringVarP(&orderDir, "order_dir", "d", "", "order direction (asc/desc)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
