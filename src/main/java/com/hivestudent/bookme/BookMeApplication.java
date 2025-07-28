package com.hivestudent.bookme;

import com.hivestudent.bookme.services.GoogleService;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableAsync;

@SpringBootApplication
@EnableAsync
public class BookMeApplication {

    public static void main(String[] args) {
        var context = SpringApplication.run(BookMeApplication.class, args);
//        context.getBean(GoogleService.class).processGoogleToken();
//        var test = context.getBean(GoogleService.class).generateGoogleJwtToken();
//        System.out.println(test);
    }

}
