package client

var _ Client = (*MockClient)(nil)

var _ Client = (*MockRenamedClient)(nil)

var _ Client = (*MockClientNoCmdComment)(nil)
