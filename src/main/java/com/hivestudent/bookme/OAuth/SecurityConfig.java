package com.hivestudent.bookme.OAuth;

import com.hivestudent.bookme.entities.Role;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpMethod;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

@Configuration
@EnableWebSecurity
public class SecurityConfig {

    @Bean
    SecurityFilterChain securityFilterChain(HttpSecurity http, JwtFilter jwtFilter) throws Exception {

        http
                .sessionManagement(c -> c.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                .csrf(c -> c.disable())
                .cors(cors -> cors.disable())
                .authorizeHttpRequests(request -> request
                        .requestMatchers(HttpMethod.POST, "/reservation").authenticated()
                        .requestMatchers(HttpMethod.GET, "/reservation").authenticated()
                        .requestMatchers(HttpMethod.GET, "/reservation/cancel/**").authenticated()
//                        .requestMatchers(HttpMethod.DELETE, "/reservation/cancel/**").hasRole(Role.STAFF.name())
                        .anyRequest().permitAll())
//                .oauth2Login(Customizer.withDefaults());
        .addFilterBefore(jwtFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }
}
