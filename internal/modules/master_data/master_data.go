package masterdata

import "context"

type MasterDataProvider interface {
	GetProvinces(ctx context.Context) ([]Region, error)
	GetRegenciesByProvince(ctx context.Context, provinceRegionCode string) ([]Region, error)
}

type MasterData struct {
	idRegionRepo IndonesiaRegionRepository
}

func NewMasterData(idRegionRepo IndonesiaRegionRepository) *MasterData {
	return &MasterData{idRegionRepo: idRegionRepo}
}

func (md *MasterData) GetProvinces(ctx context.Context) ([]Region, error) {
	return md.idRegionRepo.GetAllProvince(ctx)
}

func (md *MasterData) GetRegenciesByProvince(ctx context.Context, provinceRegionCode string) ([]Region, error) {
	return md.idRegionRepo.GetAllRegencyByProvinceRegionCode(ctx, provinceRegionCode)
}
