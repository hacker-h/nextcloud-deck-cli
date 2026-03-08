package cli

func runCapabilities(rt *runtime, args []string) error {
	data, err := rt.client.GetCapabilities(rt.ctx)
	if err != nil {
		return err
	}
	return printJSON(rt.stdout, data)
}
