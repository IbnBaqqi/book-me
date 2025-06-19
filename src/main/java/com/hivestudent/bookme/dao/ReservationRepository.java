package com.hivestudent.bookme.dao;

import com.hivestudent.bookme.entities.Reservation;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.LocalDateTime;

public interface ReservationRepository extends JpaRepository<Reservation, Long> {

//  Query to check if a slot already been reserved
  @Query(value = """
  SELECT EXISTS (
    SELECT 1
    FROM reservations
    WHERE room_id = :roomId
      AND start_time < :endTime
      AND end_time > :startTime
  )
""", nativeQuery = true)
  int existsOverlapping(@Param("roomId") Long roomId,
                            @Param("startTime") LocalDateTime startTime,
                            @Param("endTime") LocalDateTime endTime);

}