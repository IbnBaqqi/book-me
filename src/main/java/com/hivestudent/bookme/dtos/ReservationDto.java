package com.hivestudent.bookme.dtos;

import lombok.Data;

import java.time.LocalDateTime;

@Data
public class ReservationDto {
    private Long id;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private UserDto createdBy;

}
