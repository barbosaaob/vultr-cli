// Package billing provides the billing commands for the CLI
package billing

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vultr/govultr/v3"
	"github.com/vultr/vultr-cli/v3/cmd/printer"
	"github.com/vultr/vultr-cli/v3/cmd/utils"
	"github.com/vultr/vultr-cli/v3/pkg/cli"
)

var (
	long    = `Get all available commands for billing`
	example = `
	# Full example
	vultr-cli billing
	`

	historyLong    = `Get all available commands for billing history`
	historyExample = `
	# Full example
	vultr-cli billing history

	# Shortened with alias commands
	vultr-cli billing h
	`

	historyListLong    = `Retrieve a list of all billing history on your account`
	historyListExample = `
	# Full example
	vultr-cli billing history list

	# Shortened with alias commands
	vultr-cli billing h l
	`

	invoicesLong    = `Get all available commands for billing invoices`
	invoicesExample = `
	# Full example
	vultr-cli billing invoice

	# Shortened with alias commands
	vultr-cli billing i
	`

	invoiceListLong    = `Retrieve a list of all invoices on your account`
	invoiceListExample = `
	# Full example
	vultr-cli billing invoice list

	# Shortened with alias commands
	vultr-cli billing i l
	`

	invoiceGetLong    = `Get a specific invoice on your account`
	invoiceGetExample = `
	# Full example
	vultr-cli billing invoice get 123456

	# Shortened with alias commands
	vultr-cli billing i g 123456
	`

	invoiceItemsListLong    = `Retrieve a list of invoice items from a specific invoice on your account`
	invoiceItemsListExample = `
	# Full example
	vultr-cli billing invoice items 123456

	# Shortened with alias commands
	vultr-cli billing i i 123456
	`
)

func NewCmdBilling(base *cli.Base) *cobra.Command {
	o := &options{Base: base}

	cmd := &cobra.Command{
		Use:     "billing",
		Short:   "display billing information",
		Long:    long,
		Example: example,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			utils.SetOptions(o.Base, cmd, args)
			if !o.Base.HasAuth {
				return errors.New(utils.APIKeyError)
			}
			return nil
		},
	}

	// Invoice
	invoice := &cobra.Command{
		Use:     "invoice",
		Aliases: []string{"i"},
		Short:   "display invoice information",
		Long:    invoicesLong,
		Example: invoicesExample,
	}

	// Invoice List
	invoicesList := &cobra.Command{
		Use:     "list",
		Short:   "list billing invoices",
		Aliases: []string{"l"},
		Long:    invoiceListLong,
		Example: invoiceListExample,
		Run: func(cmd *cobra.Command, args []string) {
			o.Base.Options = utils.GetPaging(cmd)
			invs, meta, err := o.listInvoices()
			if err != nil {
				printer.Error(fmt.Errorf("error retrieving billing invoice list : %v", err))
				os.Exit(1)
			}
			data := &BillingInvoicesPrinter{Invoices: invs, Meta: meta}
			o.Base.Printer.Display(data, err)
		},
	}

	invoicesList.Flags().StringP("cursor", "c", "", "(optional) Cursor for paging.")
	invoicesList.Flags().IntP(
		"per-page",
		"p",
		utils.PerPageDefault,
		"(optional) Number of items requested per page. Default is 100 and Max is 500.",
	)

	// Invoice Get
	invoiceGet := &cobra.Command{
		Use:     "get",
		Short:   "get invoice",
		Aliases: []string{"g"},
		Long:    invoiceGetLong,
		Example: invoiceGetExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("please provide an invoice ID")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			inv, err := o.get()
			if err != nil {
				printer.Error(fmt.Errorf("error getting invoice : %v", err))
				os.Exit(1)
			}

			data := &BillingInvoicePrinter{Invoice: *inv}
			o.Base.Printer.Display(data, err)
		},
	}

	// Invoice Items List
	invoiceItemsList := &cobra.Command{
		Use:     "items <INVOICE_ID>",
		Short:   "list invoice items",
		Aliases: []string{"i"},
		Long:    invoiceItemsListLong,
		Example: invoiceItemsListExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("please provide an invoice ID")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			o.Base.Options = utils.GetPaging(cmd)
			id, errConv := strconv.Atoi(args[0])
			if errConv != nil {
				printer.Error(fmt.Errorf("error converting invoice item id : %v", errConv))
				os.Exit(1)
			}

			o.InvoiceItemID = id

			items, meta, err := o.listInvoiceItems()
			if err != nil {
				printer.Error(fmt.Errorf("error retrieving billing invoice item list : %v", err))
				os.Exit(1)
			}
			data := &BillingInvoiceItemsPrinter{InvoiceItems: items, Meta: meta}
			o.Base.Printer.Display(data, err)
		},
	}

	invoiceItemsList.Flags().StringP("cursor", "c", "", "(optional) Cursor for paging.")
	invoiceItemsList.Flags().IntP(
		"per-page",
		"p",
		utils.PerPageDefault,
		fmt.Sprintf("(optional) Number of items requested per page. Default is %d and Max is 500.", utils.PerPageDefault),
	)

	invoice.AddCommand(
		invoicesList,
		invoiceGet,
		invoiceItemsList,
	)

	// History
	history := &cobra.Command{
		Use:     "history",
		Aliases: []string{"h"},
		Short:   "display billing history information",
		Long:    historyLong,
		Example: historyExample,
	}

	// History List
	historyList := &cobra.Command{
		Use:     "list",
		Short:   "list billing history",
		Aliases: []string{"l"},
		Long:    historyListLong,
		Example: historyListExample,
		Run: func(cmd *cobra.Command, args []string) {
			o.Base.Options = utils.GetPaging(cmd)
			hs, meta, err := o.listHistory()
			if err != nil {
				printer.Error(fmt.Errorf("error retrieving billing history list : %v", err))
				os.Exit(1)
			}
			data := &BillingHistoryPrinter{Billing: hs, Meta: meta}
			o.Base.Printer.Display(data, err)
		},
	}

	historyList.Flags().StringP("cursor", "c", "", "(optional) Cursor for paging.")
	historyList.Flags().IntP(
		"per-page",
		"p",
		utils.PerPageDefault,
		"(optional) Number of items requested per page. Default is 100 and Max is 500.",
	)

	history.AddCommand(
		historyList,
	)

	cmd.AddCommand(
		history,
		invoice,
	)

	return cmd
}

type options struct {
	Base          *cli.Base
	InvoiceItemID int
}

func (b *options) listHistory() ([]govultr.History, *govultr.Meta, error) {
	hs, meta, _, err := b.Base.Client.Billing.ListHistory(b.Base.Context, b.Base.Options)
	return hs, meta, err
}

func (b *options) get() (*govultr.Invoice, error) {
	inv, _, err := b.Base.Client.Billing.GetInvoice(b.Base.Context, b.Base.Args[0])
	return inv, err
}

func (b *options) listInvoices() ([]govultr.Invoice, *govultr.Meta, error) {
	invs, meta, _, err := b.Base.Client.Billing.ListInvoices(b.Base.Context, b.Base.Options)
	return invs, meta, err
}

func (b *options) listInvoiceItems() ([]govultr.InvoiceItem, *govultr.Meta, error) {
	items, meta, _, err := b.Base.Client.Billing.ListInvoiceItems(b.Base.Context, b.InvoiceItemID, b.Base.Options)
	return items, meta, err
}
