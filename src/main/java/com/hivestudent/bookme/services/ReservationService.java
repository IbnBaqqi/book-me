package com.hivestudent.bookme.services;

import com.hivestudent.bookme.ReservationMapper;
import com.hivestudent.bookme.dao.ReservationRepository;
import com.hivestudent.bookme.dao.RoomRepository;
import com.hivestudent.bookme.dtos.CreateReservationRequest;
import com.hivestudent.bookme.dtos.ReservationDto;
import com.hivestudent.bookme.entities.Reservation;
import com.hivestudent.bookme.entities.ReservationStatus;
import com.hivestudent.bookme.entities.User;
import lombok.AllArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;

@Service
@AllArgsConstructor
public class ReservationService {

    private final ReservationRepository reservationRepository;
    private final RoomRepository roomRepository;
    private final ReservationMapper reservationMapper;

    //get current User
    public ReservationDto createReservation(CreateReservationRequest request, User currentUser) {

        var room = roomRepository.findById(request.getRoomId()).orElseThrow();

        LocalDateTime start = LocalDateTime.of(request.getDate(), request.getStartTime());
        LocalDateTime end = LocalDateTime.of(request.getDate(), request.getEndTime());

        if (start.isBefore(LocalDateTime.now())) {
            throw new IllegalArgumentException("You can't book past times");
        }

        var overlap = reservationRepository.existsOverlapping(room.getId(), start, end) > 0;

        if (overlap) {
            throw new IllegalArgumentException("This time slot is already booked");
        }

        Reservation reservation = new Reservation();
        reservation.setRoom(room);
        reservation.setCreatedBy(currentUser); // @Todo replace with current user later
        reservation.setStartTime(start);
        reservation.setEndTime(end);
        reservation.setStatus(ReservationStatus.RESERVED); // handled in Java, VARCHAR in DB

        reservationRepository.save(reservation);

        return reservationMapper.toDto(reservation);
    }

}
