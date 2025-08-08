package com.hivestudent.bookme.services;

import jakarta.mail.MessagingException;
import jakarta.mail.internet.MimeMessage;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.io.ClassPathResource;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.MimeMessageHelper;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.io.InputStream;
import java.io.UnsupportedEncodingException;
import java.nio.charset.StandardCharsets;

@Service
@RequiredArgsConstructor
public class EmailServiceImpl implements EmailService{

    private final JavaMailSender emailSender;

    @Value("${spring.from.email}")
    private String fromEmail;

    @Override
    @Async
    public void sendConfirmation(String email, String room, String startTime, String endTime) throws MessagingException{

        MimeMessage mimeMessage = emailSender.createMimeMessage();
        MimeMessageHelper helper = new MimeMessageHelper(mimeMessage, true, "UTF-8");

        try {
            helper.setFrom(fromEmail, "BookMe");
        } catch (UnsupportedEncodingException e) {
            throw new RuntimeException("Failed to set Sender Name");
        }

        var msgContent = getBookingEmailBody(room, startTime, endTime);

        helper.setTo(email);
        helper.setSubject("Meeting Room Confirmation");
        helper.setText(msgContent, true);

        emailSender.send(mimeMessage);
    }

    private static String getBookingEmailBody(String roomName, String startTime, String endTime) {
        try {
            ClassPathResource resource = new ClassPathResource("booking_email_template.html");
            try (InputStream inputStream = resource.getInputStream()) {
                String template = new String(inputStream.readAllBytes(), StandardCharsets.UTF_8);
                return template
                        .replace("__ROOM_NAME__", roomName)
                        .replace("__START_TIME__", startTime)
                        .replace("__END_TIME__", endTime);
            }
        } catch (IOException e) {
            return String.format("Hi, the %s meeting room has been reserved for you from %s to %s.", roomName, startTime, endTime);
        }
    }
}
