package com.hivestudent.bookme.dao;

import com.hivestudent.bookme.entities.Reservation;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

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
  long existsOverlapping(@Param("roomId") Long roomId,
                            @Param("startTime") LocalDateTime startTime,
                            @Param("endTime") LocalDateTime endTime);

  @Query("""
SELECT r FROM Reservation r
JOIN FETCH r.room
JOIN FETCH r.createdBy
WHERE r.startTime >= :startDate
  AND r.endTime <= :endDate
ORDER BY r.room.id, r.startTime
""")
  List<Reservation> findAllBetweenDates(
          @Param("startDate") LocalDateTime startDate,
          @Param("endDate") LocalDateTime endDate
  );

  @Query(value = """
SELECT SUM(TIMESTAMPDIFF(MINUTE, r.start_time, r.end_time))
FROM reservations r
WHERE r.user_id = :userId
  AND DATE(r.start_time) = :date
  AND r.status = 'RESERVED'
""", nativeQuery = true)
  Integer getTotalReservedMinutesForDate(@Param("userId") Long userId,
                                         @Param("date") LocalDate date);

}