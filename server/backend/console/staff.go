package console

import (
	"fmt"

	"euphoria.leet.nu/heim/console"
	"euphoria.leet.nu/heim/proto/security"
)

func init() {
	register("grant-staff", "Grant staff privileges to an account.", &grantStaff{})
	register("revoke-staff", "Remove staff privileges from an account.", &revokeStaff{})
}

type grantStaff struct {
	handlerBase
	Account     string `usage:"Account to modify." cli:",arg,required"`
	KMSType     string `usage:"KMS to use for privileged actions." cli:"kms-type,arg,required"`
	Credentials string `usage:"Credentials for the KMS." cli:",arg"`
}

func (g *grantStaff) Run(env console.CLIEnv) error {
	ctx := env.Context()

	account, err := g.cli.resolveAccount(ctx, g.Account)
	if err != nil {
		return err
	}

	kmsType := security.KMSType(g.KMSType)
	kmsCred, err := kmsType.KMSCredential()
	if err != nil {
		return err
	}

	if g.Credentials == "" && kmsType == security.LocalKMSType {
		mockKMS, ok := g.cli.kms.(security.MockKMS)
		if !ok {
			return fmt.Errorf("this backend does not support KMS type %s", kmsType)
		}
		kmsCred = mockKMS.KMSCredential()
	} else {
		if g.Credentials == "" {
			g.Credentials, err = env.ReadPassword("KMS credentials: ")
			if err != nil {
				return err
			}
		}
		if err := kmsCred.UnmarshalJSON([]byte(g.Credentials)); err != nil {
			return err
		}
	}

	env.Printf("Granting staff capability to account %s\n", account.ID())
	return g.cli.backend.AccountManager().GrantStaff(ctx, account.ID(), kmsCred)
}

type revokeStaff struct {
	handlerBase
	Account string `usage:"Account to modify." cli:",arg,required"`
}

func (r *revokeStaff) Run(env console.CLIEnv) error {
	ctx := env.Context()

	account, err := r.cli.resolveAccount(ctx, r.Account)
	if err != nil {
		return err
	}

	env.Printf("Revoking staff capability from %s\n", account.ID())
	if !account.IsStaff() {
		env.Printf("NOTE: This account isn't currently holding a staff capability\n")
	}
	return r.cli.backend.AccountManager().RevokeStaff(ctx, account.ID())
}
