package blobstore

import (
	"fmt"
	"path"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

const (
	BlobstoreTypeDummy = "dummy"
	BlobstoreTypeLocal = "local"
)

type Provider struct {
	fs        system.FileSystem
	runner    system.CmdRunner
	configDir string
	uuidGen   boshuuid.Generator
	logger    boshlog.Logger
}

func NewProvider(
	fs system.FileSystem,
	runner system.CmdRunner,
	configDir string,
	logger boshlog.Logger,
) Provider {
	return Provider{
		uuidGen:   boshuuid.NewGenerator(),
		fs:        fs,
		runner:    runner,
		configDir: configDir,
		logger:    logger,
	}
}

func (p Provider) Get(storeType string, options map[string]interface{}) (Blobstore, error) {
	var blobstore Blobstore

	switch storeType {
	case BlobstoreTypeDummy:
		blobstore = newDummyBlobstore()

	case BlobstoreTypeLocal:
		blobstore = NewLocalBlobstore(
			p.fs,
			p.uuidGen,
			options,
		)

	default:
		blobstore = NewExternalBlobstore(
			storeType,
			options,
			p.fs,
			p.runner,
			p.uuidGen,
			path.Join(p.configDir, fmt.Sprintf("blobstore-%s.json", storeType)),
		)
	}

	blobstore = NewRetryableBlobstore(NewDigestVerifiableBlobstore(blobstore), 3, p.logger)

	err := blobstore.Validate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Validating blobstore")
	}

	return blobstore, nil
}
