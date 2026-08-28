package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dlvandenberg/printer-ledger/internal/domain"
)

var _ domain.SettingsRepository = &Store{}

func (s *Store) Settings(ctx context.Context) (domain.Settings, error) {
	row := s.q().QueryRowContext(ctx, `
SELECT kwh_price_cents, machine_hourly_rate_cents, printer_purchase_cost_cents,
       default_margin_hundredths, min_margin_hundredths
FROM settings
WHERE id = 1`)

	settings, err := scanSettings(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, fmt.Errorf("read settings: %w", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("read settings: %w", err)
	}

	rates, err := s.powerRates(ctx)
	if err != nil {
		return domain.Settings{}, err
	}
	settings.PowerRates = rates
	return settings, nil
}

func (s *Store) SaveSettings(ctx context.Context, settings domain.Settings) error {
	_, err := s.q().ExecContext(ctx, `
UPDATE settings
SET kwh_price_cents             = ?,
    machine_hourly_rate_cents   = ?,
    printer_purchase_cost_cents = ?,
    default_margin_hundredths   = ?,
    min_margin_hundredths       = ?
WHERE id = 1`,
		int64(settings.KwhPrice), int64(settings.MachineHourlyRate),
		int64(settings.PrinterPurchaseCost), int64(settings.DefaultMargin),
		int64(settings.MinMargin))
	if err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	for _, rate := range settings.PowerRates {
		_, err := s.q().ExecContext(ctx, `
INSERT INTO power_rates (filament_type, kwh_per_hour, measured)
VALUES (?, ?, ?)
ON CONFLICT (filament_type) DO UPDATE SET kwh_per_hour = excluded.kwh_per_hour,
                                          measured     = excluded.measured`,
			string(rate.FilamentType), float64(rate.KwhPerHour), rate.Measured)
		if err != nil {
			return fmt.Errorf("save %s power rate: %w", rate.FilamentType, err)
		}
	}
	return nil
}

// seedSettings runs on every open, so a ledger created before a Filament Type
// existed gains its rate rather than costing that material's energy at zero
// (ADR-0008, ADR-0010).
func (s *Store) seedSettings(ctx context.Context) error {
	seeded := domain.DefaultSettings()
	_, err := s.q().ExecContext(ctx, `
INSERT OR IGNORE INTO settings (id, kwh_price_cents, machine_hourly_rate_cents,
                                printer_purchase_cost_cents, default_margin_hundredths,
                                min_margin_hundredths)
VALUES (1, ?, ?, ?, ?, ?)`,
		int64(seeded.KwhPrice), int64(seeded.MachineHourlyRate),
		int64(seeded.PrinterPurchaseCost), int64(seeded.DefaultMargin),
		int64(seeded.MinMargin))
	if err != nil {
		return fmt.Errorf("seed settings: %w", err)
	}

	for _, rate := range seeded.PowerRates {
		_, err := s.q().ExecContext(ctx, `
INSERT OR IGNORE INTO power_rates (filament_type, kwh_per_hour, measured)
VALUES (?, ?, ?)`,
			string(rate.FilamentType), float64(rate.KwhPerHour), rate.Measured)
		if err != nil {
			return fmt.Errorf("seed %s power rate: %w", rate.FilamentType, err)
		}
	}
	return nil
}

func (s *Store) powerRates(ctx context.Context) ([]domain.PowerRate, error) {
	rows, err := s.q().QueryContext(ctx, `
SELECT filament_type, kwh_per_hour, measured FROM power_rates ORDER BY filament_type`)
	if err != nil {
		return nil, fmt.Errorf("read power rates: %w", err)
	}
	defer rows.Close()

	var rates []domain.PowerRate
	for rows.Next() {
		rate, err := scanPowerRate(rows)
		if err != nil {
			return nil, fmt.Errorf("read power rates: %w", err)
		}
		rates = append(rates, rate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read power rates: %w", err)
	}
	return rates, nil
}

func scanSettings(row scanner) (domain.Settings, error) {
	var kwhPrice, machineHourlyRate, printerPurchaseCost, defaultMargin, minMargin int64
	if err := row.Scan(&kwhPrice, &machineHourlyRate, &printerPurchaseCost,
		&defaultMargin, &minMargin); err != nil {
		return domain.Settings{}, err
	}
	return domain.Settings{
		KwhPrice:            domain.Cents(kwhPrice),
		MachineHourlyRate:   domain.Cents(machineHourlyRate),
		PrinterPurchaseCost: domain.Cents(printerPurchaseCost),
		DefaultMargin:       domain.Percent(defaultMargin),
		MinMargin:           domain.Percent(minMargin),
	}, nil
}

func scanPowerRate(row scanner) (domain.PowerRate, error) {
	var (
		filamentType string
		kwhPerHour   float64
		measured     bool
	)
	if err := row.Scan(&filamentType, &kwhPerHour, &measured); err != nil {
		return domain.PowerRate{}, err
	}
	return domain.PowerRate{
		FilamentType: domain.FilamentType(filamentType),
		KwhPerHour:   domain.KwhPerHour(kwhPerHour),
		Measured:     measured,
	}, nil
}
