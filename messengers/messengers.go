package messengers

func IsBang(msg any) bool {
	bang, ok := msg.(string)

	return ok && bang == "bang"
}
