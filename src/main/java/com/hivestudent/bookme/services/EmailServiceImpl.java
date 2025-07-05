package com.hivestudent.bookme.services;

import jakarta.mail.MessagingException;
import jakarta.mail.internet.MimeMessage;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.MimeMessageHelper;
import org.springframework.stereotype.Service;

import java.io.UnsupportedEncodingException;

@Service
@RequiredArgsConstructor
public class EmailServiceImpl implements EmailService{

    private final JavaMailSender emailSender;

    @Value("${spring.from.email}")
    private String fromEmail;

    @Override
    public void sendEmail(String email) throws MessagingException {

        MimeMessage mimeMessage = emailSender.createMimeMessage();
        MimeMessageHelper helper = new MimeMessageHelper(mimeMessage, true, "UTF-8");

        try {
            helper.setFrom(fromEmail, "Book-me Test");
        } catch (UnsupportedEncodingException e) {
            throw new RuntimeException("Failed to set Sender Name");
        }
        helper.setTo(email);
        helper.setSubject("Meeting Room Confirmation");
        helper.setText("Your Meeting Reservation has been booked");

        emailSender.send(mimeMessage);
    }
}
