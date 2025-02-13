package ovirtclient

import (
	"fmt"
	ovirtsdk "github.com/ovirt/go-ovirt"
	"sync"
)

func (o *oVirtClient) MoveDiskToStorageDomain(
	diskID DiskID,
	storageDomainID StorageDomainID,
	retries ...RetryStrategy) (result Disk, err error) {
	retries = defaultRetries(retries, defaultReadTimeouts(o))
	progress, err := o.StartMoveDiskToStorageDomain(diskID, storageDomainID, retries...)
	if err != nil {
		return nil, err
	}

	return progress.Wait()
}

func (o *oVirtClient) StartMoveDiskToStorageDomain(
	diskID DiskID,
	storageDomainID StorageDomainID,
	retries ...RetryStrategy) (DiskUpdate, error) {
	retries = defaultRetries(retries, defaultWriteTimeouts(o))
	correlationID := fmt.Sprintf("template_disk_copy_%s", generateRandomID(5, o.nonSecureRandom))
	sdkStorageDomain := ovirtsdk.NewStorageDomainBuilder().Id(string(storageDomainID))
	storageDomain, _ := o.GetStorageDomain(storageDomainID)
	disk, _ := o.GetDisk(diskID)

	err := retry(
		fmt.Sprintf("moving disk %s to storage domain %s", diskID, storageDomainID),
		o.logger,
		retries,
		func() error {
			_, err := o.conn.
				SystemService().
				DisksService().
				DiskService(string(diskID)).
				Move().
				StorageDomain(sdkStorageDomain.MustBuild()).
				Query("correlation_id", correlationID).
				Send()

			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return &storageDomainDiskWait{
		client:        o,
		disk:          disk,
		storageDomain: storageDomain,
		correlationID: correlationID,
		lock:          &sync.Mutex{},
	}, nil
}

func (o *mockClient) MoveDiskToStorageDomain(
	diskID DiskID,
	storageDomainID StorageDomainID,
	retries ...RetryStrategy) (result Disk, err error) {
	panic("not implemented")
}

func (o *mockClient) StartMoveDiskToStorageDomain(
	diskID DiskID,
	storageDomainID StorageDomainID,
	retries ...RetryStrategy) (DiskUpdate, error) {
	panic("not implemented")
}
