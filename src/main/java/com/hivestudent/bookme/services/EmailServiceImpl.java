package com.hivestudent.bookme.services;

import jakarta.mail.MessagingException;
import jakarta.mail.internet.MimeMessage;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.MimeMessageHelper;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.io.UnsupportedEncodingException;
import java.nio.file.Files;
import java.nio.file.Paths;

@Service
@RequiredArgsConstructor
public class EmailServiceImpl implements EmailService{

    private final JavaMailSender emailSender;

    @Value("${spring.from.email}")
    private String fromEmail;

    @Override
    @Async
    public void sendConfirmation(String email, String room, String date) throws MessagingException{

        MimeMessage mimeMessage = emailSender.createMimeMessage();
        MimeMessageHelper helper = new MimeMessageHelper(mimeMessage, true, "UTF-8");

        try {
            helper.setFrom(fromEmail, "BookMe");
        } catch (UnsupportedEncodingException e) {
            throw new RuntimeException("Failed to set Sender Name");
        }

//        var htmlContent = getBookingEmailBody(room, date);

        var msgContent = createMessage(room, date);

        helper.setTo(email);
        helper.setSubject("Meeting Room Confirmation");
        helper.setText(msgContent, true);

        emailSender.send(mimeMessage);
    }

//    public static String getBookingEmailBody(String roomSize, String dateTime) throws IOException {
//        String templatePath = "src/main/resources/booking_email_template.html";
//        String template = Files.readString(Paths.get(templatePath));
//
//        return template
//                .replace("${roomSize}", roomSize)
//                .replace("${dateTime}", dateTime);
//    }

    private static String createMessage(String roomSize, String dateTime) {
        return String.format("Hi, the %s meeting room has been reserved for you from %s.", roomSize, dateTime);
    }
}
