package com.hivestudent.bookme;

import com.hivestudent.bookme.dtos.ReservationDto;
import com.hivestudent.bookme.dtos.UpdateReservationRequest;
import com.hivestudent.bookme.entities.Reservation;
import org.mapstruct.Mapper;
import org.mapstruct.MappingTarget;

@Mapper(componentModel = "spring")
public interface ReservationMapper {
    ReservationDto toDto(Reservation reservation);
    void update(UpdateReservationRequest request, @MappingTarget Reservation reservation);
}
