-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS master_data_indonesia_provinces (
   region_code VARCHAR(2) NOT NULL PRIMARY KEY,
   name VARCHAR(26) NOT NULL
);

INSERT INTO master_data_indonesia_provinces(region_code,name) VALUES
 ('11','ACEH')
,('12','SUMATERA UTARA')
,('13','SUMATERA BARAT')
,('14','RIAU')
,('15','JAMBI')
,('16','SUMATERA SELATAN')
,('17','BENGKULU')
,('18','LAMPUNG')
,('19','KEPULAUAN BANGKA BELITUNG')
,('21','KEPULAUAN RIAU')
,('31','DKI JAKARTA')
,('32','JAWA BARAT')
,('33','JAWA TENGAH')
,('34','DAERAH ISTIMEWA YOGYAKARTA')
,('35','JAWA TIMUR')
,('36','BANTEN')
,('51','BALI')
,('52','NUSA TENGGARA BARAT')
,('53','NUSA TENGGARA TIMUR')
,('61','KALIMANTAN BARAT')
,('62','KALIMANTAN TENGAH')
,('63','KALIMANTAN SELATAN')
,('64','KALIMANTAN TIMUR')
,('65','KALIMANTAN UTARA')
,('71','SULAWESI UTARA')
,('72','SULAWESI TENGAH')
,('73','SULAWESI SELATAN')
,('74','SULAWESI TENGGARA')
,('75','GORONTALO')
,('76','SULAWESI BARAT')
,('81','MALUKU')
,('82','MALUKU UTARA')
,('91','PAPUA')
,('92','PAPUA BARAT')
,('93','PAPUA SELATAN')
,('94','PAPUA TENGAH')
,('95','PAPUA PEGUNUNGAN');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS master_data_indonesia_provinces;
-- +goose StatementEnd
