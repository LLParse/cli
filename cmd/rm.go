package cmd

import (
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

var (
	rmTypes = []string{"service", "container", "stack", "host", "volume"}
)

func RmCommand() cli.Command {
	return cli.Command{
		Name:        "rm",
		Usage:       "Delete " + strings.Join(rmTypes, ", "),
		Description: "\nDeletes resources by ID or name in the current $RANCHER_ENVIRONMENT.  Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher rm 1s70\n\t$ rancher --env 1a5 rm stackName/serviceName \n",
		ArgsUsage:   "[ID NAME...]",
		Action:      deleteResources,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "type",
				Usage: "Restrict delete to specific types",
				Value: &cli.StringSlice{},
			},
			cli.BoolFlag{
				Name:  "stop,s",
				Usage: "Stop or deactivate resource first if needed before deleting",
			},
		},
	}
}

func deleteResources(ctx *cli.Context) error {
	return forEachResource(ctx, rmTypes, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		if ctx.Bool("stop") {
			action, err := pickAction(resource, "stop", "deactivate")
			if err == nil {
				w, err := NewWaiter(ctx)
				if err != nil {
					return "", err
				}
				if err := c.Action(resource.Type, action, resource, nil, resource); err != nil {
					return "", err
				}
				if err := w.Add(resource.Id).Wait(); err != nil {
					return "", err
				}
			}
		}
		err := c.Delete(resource)
		if v, ok := err.(*client.ApiError); ok && v.StatusCode == 405 {
			action, err := pickAction(resource, "stop", "deactivate")
			if err == nil {
				fmt.Printf("error: Must call %s on %s %s before removing\n", action, resource.Type, resource.Id)
			}
		}
		return resource.Id, err
	})
}
