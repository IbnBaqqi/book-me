package com.hivestudent.bookme.dtos;

import lombok.Data;
import org.springframework.format.annotation.DateTimeFormat;

import java.time.LocalDate;
import java.time.LocalTime;

@Data
public class CreateReservationRequest {

//    @Todo change room id to byte
    private Long roomId;

    @DateTimeFormat(iso = DateTimeFormat.ISO.DATE)
    private LocalDate date; // e.g. 2025-06-17

    @DateTimeFormat(pattern = "HH:mm")
    private LocalTime startTime; // e.g. 06:00

    @DateTimeFormat(pattern = "HH:mm")
    private LocalTime endTime; // e.g. 07:00


}
