package api

import "context"

func (a *ApiService) MailcowGetAlias(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/alias/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetAliasDomain(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/alias-domain/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetDomain(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/domain/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetMailbox(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/mailbox/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetDkim(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/dkim/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetSyncjob(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/syncjobs/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetOAuth2Client(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/oauth2-client/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetOAuth2Clients(ctx context.Context) ApiMailcowGetAllRequest {
	return ApiMailcowGetAllRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/oauth2-client/all",
	}
}

func (a *ApiService) MailcowGetDomainAdminAll(ctx context.Context) ApiMailcowGetAllRequest {
	return ApiMailcowGetAllRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/domain-admin/all",
	}
}

func (a *ApiService) MailcowGetIdentityProviderKeycloak(ctx context.Context) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/identity-provider",
	}
}

func (a *ApiService) MailcowGetBcc(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/bcc/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetBccAll(ctx context.Context) ApiMailcowGetAllRequest {
	return ApiMailcowGetAllRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/bcc/all",
	}
}

func (a *ApiService) MailcowGetRelayhost(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/relayhost/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetTlsPolicyMap(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/tls-policy-map/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetTlsPolicyMapAll(ctx context.Context) ApiMailcowGetAllRequest {
	return ApiMailcowGetAllRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/tls-policy-map/all",
	}
}

func (a *ApiService) MailcowGetFwdhostAll(ctx context.Context) ApiMailcowGetAllRequest {
	return ApiMailcowGetAllRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/fwdhost/all",
	}
}

func (a *ApiService) MailcowGetRecipientMap(ctx context.Context, id string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/recipient_map/{id}",
		id:         id,
	}
}

func (a *ApiService) MailcowGetResourceAll(ctx context.Context) ApiMailcowGetAllRequest {
	return ApiMailcowGetAllRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   "/api/v1/get/resource/all",
	}
}

func (a *ApiService) MailcowGetDomainPolicy(ctx context.Context, domain string, endpoint string) ApiMailcowGetRequest {
	return ApiMailcowGetRequest{
		ApiService: a,
		ctx:        ctx,
		endpoint:   endpoint,
		id:         domain,
	}
}
