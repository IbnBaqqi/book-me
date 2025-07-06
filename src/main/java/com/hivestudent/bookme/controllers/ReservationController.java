package com.hivestudent.bookme.controllers;

import com.hivestudent.bookme.Auth.OAuthService;
import com.hivestudent.bookme.dtos.CreateReservationRequest;
import com.hivestudent.bookme.dtos.ReservationDto;
import com.hivestudent.bookme.dtos.ReservedDto;
import com.hivestudent.bookme.dtos.UpdateReservationRequest;
import com.hivestudent.bookme.entities.User;
import com.hivestudent.bookme.services.ReservationService;
import jakarta.mail.MessagingException;
import jakarta.validation.Valid;
import lombok.AllArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.time.LocalDate;
import java.util.List;
import java.util.Map;

@AllArgsConstructor
@RestController
@RequestMapping("/reservation")
public class ReservationController {

    private final ReservationService reservationService;
    private final OAuthService oAuthService;

    @PostMapping
    public ResponseEntity<ReservationDto> create(@RequestBody @Valid CreateReservationRequest request ) throws MessagingException, IOException {

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

    @PutMapping("/{id}")
    public ResponseEntity<ReservationDto> update(@PathVariable Long id, @Valid @RequestBody UpdateReservationRequest request) {
        var reservation = reservationService.updateReservation(id, request);
        return ResponseEntity.ok(reservation);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> cancelReservation(@PathVariable Long id) {
        reservationService.cancelReservation(id);
        return ResponseEntity.noContent().build();
    }

    @ExceptionHandler(IllegalArgumentException.class)
    public ResponseEntity<Map<String, String>> handleIllegalArgumentExceptions(Exception e) {
        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(
                Map.of("error", e.getMessage())
        );
    }

}
