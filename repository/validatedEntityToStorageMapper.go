package repository

import (
	ent "barbershop-bot/entities"
	u "barbershop-bot/lib/utils"
	st "barbershop-bot/repository/storage"
)

type validatedEntityToStorageMapper struct {
	entityToStorageMapper
}

var mapToStorage validatedEntityToStorageMapper

func (v *validatedEntityToStorageMapper) barber(barber *ent.Barber) (st.Barber, error) {
	if barber.Name != "" && !u.IsValidName(barber.Name) {
		return st.Barber{}, ErrInvalidName
	}
	if barber.Phone != "" {
		if !u.IsValidPhone(barber.Phone) {
			return st.Barber{}, ErrInvalidPhone
		}
		barber.Phone = u.NormalizePhone(barber.Phone)
	}
	return v.entityToStorageMapper.barber(barber)
}
