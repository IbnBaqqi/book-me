package com.hivestudent.bookme.services;

import com.hivestudent.bookme.Auth.OAuthService;
import com.hivestudent.bookme.mapper.ReservationMapper;
import com.hivestudent.bookme.dao.ReservationRepository;
import com.hivestudent.bookme.dao.RoomRepository;
import com.hivestudent.bookme.dtos.CreateReservationRequest;
import com.hivestudent.bookme.dtos.ReservationDto;
import com.hivestudent.bookme.dtos.ReservedDto;
import com.hivestudent.bookme.dtos.UpdateReservationRequest;
import com.hivestudent.bookme.entities.Reservation;
import com.hivestudent.bookme.entities.ReservationStatus;
import com.hivestudent.bookme.entities.Role;
import com.hivestudent.bookme.entities.User;
import jakarta.mail.MessagingException;
import lombok.AllArgsConstructor;
import org.springframework.security.core.Authentication;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.time.Duration;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Service
@AllArgsConstructor
public class ReservationService {

    private final ReservationRepository reservationRepository;
    private final RoomRepository roomRepository;
    private final ReservationMapper reservationMapper;
    private final OAuthService oAuthService;
    private final EmailService emailService;


    public ReservationDto createReservation(CreateReservationRequest request, User currentUser) throws MessagingException, IOException {

        var room = roomRepository.findById(request.getRoomId()).orElseThrow();

        var start = request.getStartTime();
        var end = request.getEndTime();

        if (start.isBefore(LocalDateTime.now())) {
            throw new IllegalArgumentException("You can't book past times");
        }

        var overlap = reservationRepository.existsOverlapping(room.getId(), start, end) > 0;

        if (overlap) {
            throw new IllegalArgumentException("This time slot is already booked");
        }

//        Set max Booking Time for Student
        var maxTime = 240; // 4 hours
        var duration = Duration.between(start, end);
        if (duration.toMinutes() > maxTime && currentUser.getRole().equals(Role.STUDENT))
            throw new IllegalArgumentException("Reservation exceeds maximum allowed duration of 4 hour");

        Reservation reservation = new Reservation();
        reservation.setRoom(room);
        reservation.setCreatedBy(currentUser);
        reservation.setStartTime(start);
        reservation.setEndTime(end);
        reservation.setStatus(ReservationStatus.RESERVED);

        reservationRepository.save(reservation);

        var date = reservation.dateToEmailFormat();

        emailService.sendConfirmation(currentUser.getEmail(), room.getName(), date);

        return reservationMapper.toDto(reservation);
    }

    public List<ReservedDto> getUnavailableSlots(LocalDate start, LocalDate end, Authentication authentication) {
        var startDateTime = start.atStartOfDay();
        var endDateTime = end.plusDays(1).atStartOfDay();

        boolean isStaff = authentication.getAuthorities().stream()
                .anyMatch(a -> a.getAuthority().equals("ROLE_STAFF"));

        var reservations = reservationRepository.findAllBetweenDates(startDateTime, endDateTime);

        var grouped = reservations.stream()
                .collect(Collectors.groupingBy(res -> res.getRoom().getId()));

        List<ReservedDto> result = new ArrayList<>();

        for (var entry : grouped.entrySet()) {

            Long roomId = entry.getKey();
            List<Reservation> roomReservations = entry.getValue();
            String roomName = roomReservations.get(0).getRoom().getName();

            List<ReservedDto.Slot> slots = entry.getValue().stream()
                    .map(r -> {

                        boolean isOwner = r.getCreatedBy().getEmail().equals(authentication.getPrincipal());
                        String bookedBy = (isStaff || isOwner) ? r.getCreatedBy().getName() : null;

                                return new ReservedDto.Slot(
                                        r.getId(),
                                        r.getStartTime(),
                                        r.getEndTime(),
                                        bookedBy
                                );
                    }).toList();

            result.add(new ReservedDto(roomId, roomName, slots));
        }

        return result;
    }

    public void cancelReservation(Long id) {

        var reserved = reservationRepository.findById(id).orElse(null);
        if (reserved == null)
            throw new IllegalArgumentException("Reservation doesn't exist");

        var user = oAuthService.getCurrentUser();
        boolean isStaff = user.getRole().equals(Role.STAFF);
        boolean isOwner = reserved.getCreatedBy().getEmail().equals(user.getEmail());

        if (isStaff || isOwner) {
            reservationRepository.delete(reserved);
            reserved.setCreatedBy(null);
            reserved.setRoom(null);
        }else
            throw new IllegalArgumentException("You didn't book this slot");
    }

    public ReservationDto updateReservation(Long id, UpdateReservationRequest request) {

        var reserved = reservationRepository.findById(id).orElse(null);
        if (reserved == null)
            throw new IllegalArgumentException("Reservation doesn't exist");

        var user = oAuthService.getCurrentUser();
        boolean isOwner = reserved.getCreatedBy().getEmail().equals(user.getEmail());

        var overlap = reservationRepository.existsOverlapping(request.getRoomId(), request.getStartTime(), request.getEndTime()) > 0;
        if (overlap)
            throw new IllegalArgumentException("This time slot is already booked");

        if (isOwner) {
            reservationMapper.update(request, reserved);
            reservationRepository.save(reserved);
        }

        return reservationMapper.toDto(reserved);
    }
}
