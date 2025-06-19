package com.hivestudent.bookme.controllers;

import com.hivestudent.bookme.dtos.CreateReservationRequest;
import com.hivestudent.bookme.dtos.ReservationDto;
import com.hivestudent.bookme.services.ReservationService;
import jakarta.validation.Valid;
import lombok.AllArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@AllArgsConstructor
@RestController
@RequestMapping("/reservation")
public class ReservationController {

    private final ReservationService reservationService;

    //create reservation
    //authenticated
    //must not overlap
    @PostMapping
    public ResponseEntity<ReservationDto> create(@RequestBody @Valid CreateReservationRequest request ) {

//        User currentUser = userService.getCurrentUser(); // based on OAuth2 email

        var reserved = reservationService.createReservation(request);

        return ResponseEntity.status(HttpStatus.CREATED).body(reserved);
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<Map<String, String>> handleIllegalArgumentExceptions(Exception e) {
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(
                Map.of("error", e.getMessage())
        );
    }

}
