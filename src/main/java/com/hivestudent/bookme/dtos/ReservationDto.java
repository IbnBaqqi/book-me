package com.hivestudent.bookme.dtos;

import lombok.Data;

import java.time.LocalDateTime;

@Data
public class ReservationDto {
    // don't send createdBy if role not staff or user isn't createdBy current user
    private Long id;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private UserDto createdBy;

}
