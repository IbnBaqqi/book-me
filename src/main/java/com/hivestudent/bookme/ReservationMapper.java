package com.hivestudent.bookme;

import com.hivestudent.bookme.dtos.ReservationDto;
import com.hivestudent.bookme.entities.Reservation;
import org.mapstruct.Mapper;

@Mapper(componentModel = "spring")
public interface ReservationMapper {
    ReservationDto toDto(Reservation reservation);
}
