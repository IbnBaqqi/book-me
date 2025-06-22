package com.hivestudent.bookme.controllers;

import com.hivestudent.bookme.OAuth.OAuthService;
import com.hivestudent.bookme.dtos.CreateReservationRequest;
import com.hivestudent.bookme.dtos.ReservationDto;
import com.hivestudent.bookme.dtos.ReservedDto;
import com.hivestudent.bookme.entities.User;
import com.hivestudent.bookme.services.ReservationService;
import jakarta.validation.Valid;
import lombok.AllArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;
import java.util.Map;

@AllArgsConstructor
@RestController
@RequestMapping("/reservation")
public class ReservationController {

    private final ReservationService reservationService;
    private final OAuthService oAuthService;

    //create reservation
    //authenticated
    //must not overlap
    @PostMapping
    public ResponseEntity<ReservationDto> create(@RequestBody @Valid CreateReservationRequest request ) {

        User currentUser = oAuthService.getCurrentUser(); // based on OAuth2 email

        var reserved = reservationService.createReservation(request, currentUser);

        return ResponseEntity.status(HttpStatus.CREATED).body(reserved);
    }

    @GetMapping
    public ResponseEntity<List<ReservedDto>> getUnavailableSlots(
            @RequestParam("start") @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate start,
            @RequestParam("end") @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate end,
            Authentication authentication
    ) {
        var reserved = reservationService.getUnavailableSlots(start, end, authentication);
        return ResponseEntity.ok(reserved);
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<Map<String, String>> handleIllegalArgumentExceptions(Exception e) {
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(
                Map.of("error", e.getMessage())
        );
    }

}
