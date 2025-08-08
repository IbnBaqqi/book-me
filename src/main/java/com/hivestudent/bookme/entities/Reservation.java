package com.hivestudent.bookme.entities;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.Setter;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.time.format.TextStyle;
import java.util.Locale;

@Getter
@Setter
@Entity
@Table(name = "reservations")
public class Reservation {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    @ManyToOne
    @JoinColumn(name = "user_id")
    private User createdBy;

    @ManyToOne
    @JoinColumn(name = "room_id")
    private Room room;

    @Column(name = "start_time")
    private LocalDateTime startTime;

    @Column(name = "end_time")
    private LocalDateTime endTime;

    @Column(name = "status")
    @Enumerated(EnumType.STRING)
    private ReservationStatus status;

    @Column(name = "gcal_event_id")
    private String googleCalendarEventId;

    public String dateToEmailFormat(LocalDateTime dateTime) {

        var day = dateTime.getDayOfMonth();
        var month = dateTime.getMonth().getDisplayName(TextStyle.FULL, Locale.ENGLISH);
        var year = dateTime.getYear();

        return month + " " + day + ", " + year + " " + dateTime.format(DateTimeFormatter.ofPattern("HH:mm"));
    }

}