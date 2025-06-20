package com.hivestudent.bookme.OAuth;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.AllArgsConstructor;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

@Component
@AllArgsConstructor
public class JwtFilter extends OncePerRequestFilter {

    private final JwtService jwtService;

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain) throws ServletException, IOException {
        var authHeader = request.getHeader("Authorization");
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            filterChain.doFilter(request, response);
            return;
        }

        var token = authHeader.replace("Bearer ", "");
        if (!jwtService.validateToken(token)) {
            filterChain.doFilter(request, response);
            return;
        }
        // authentication
        var authentication = new UsernamePasswordAuthenticationToken(
                jwtService.extractEmail(token),
                null,
                null
        );

        //Attaching additional metadata about the request e.g ip address etc to the authentication object
        authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));

        //Stores current authentication to securityContextHolder, to access current user
        SecurityContextHolder.getContext().setAuthentication(authentication);

        //pass to next filter in chain
        filterChain.doFilter(request, response);
    }
}
