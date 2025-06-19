package com.hivestudent.bookme.dtos;

import com.hivestudent.bookme.entities.User;
import lombok.Data;

import java.time.LocalDateTime;

@Data
public class ReservationDto {
    private Long id;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private User createdBy;

}
