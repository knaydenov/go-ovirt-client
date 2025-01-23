package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) MoveDiskToStorageDomain(sourceId StorageDomainID, targetId StorageDomainID, diskID DiskID, retries ...RetryStrategy) (err error) {
	retries = defaultRetries(retries, defaultReadTimeouts(o))
	err = retry(
		fmt.Sprintf("movomg disk %s from %s to %s storage domain", diskID, sourceId, targetId),
		o.logger,
		retries,
		func() error {
			storageDomainGetResponse, err := o.conn.SystemService().StorageDomainsService().StorageDomainService(string(targetId)).Get().Send()
			if err != nil {
				o.logger.Infof("error getting storage domain..")
				return err
			}
			targetStorageDomain, ok := storageDomainGetResponse.StorageDomain()
			if !ok {
				o.logger.Infof("error getting storage domain..")
				return err
			}
			_, err = o.conn.SystemService().StorageDomainsService().
				StorageDomainService(string(sourceId)).DisksService().DiskService(string(diskID)).Move().StorageDomain(targetStorageDomain).Send()
			if err != nil {
				o.logger.Infof("error moving disk..")
				return err
			}

			return nil
		})
	return
}

func (m *mockClient) MoveDiskToStorageDomain(sourceId StorageDomainID, targetId StorageDomainID, diskID DiskID, _ ...RetryStrategy) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.disks[diskID]; !ok {
		return newError(ENotFound, "disk with ID %s not found", diskID)
	}

	domains := m.disks[diskID].storageDomainIDs

	// if there is only 1 domain just delete the disk
	if len(domains) == 1 {
		delete(m.disks, diskID)
		return nil
	}

	// remove the storagedomain from the disk slice
	for i, sdomain := range domains {
		if sdomain == sourceId {
			// gocritic will complain on the following line due to appendAssign, but that's legit here
			m.disks[diskID].storageDomainIDs = append(domains[:i], domains[i+1:]...) //nolint:gocritic
			return nil
		}
	}
	// add the storagedomain to the disk slice
	for i, sdomain := range domains {
		if sdomain == targetId {
			// gocritic will complain on the following line due to appendAssign, but that's legit here
			m.disks[diskID].storageDomainIDs = append(domains[:i], domains[i+1:]...) //nolint:gocritic
			return nil
		}
	}
	return newError(ENotFound, "disk %s is not found in StorageDomain %s", diskID, sourceId)
}
