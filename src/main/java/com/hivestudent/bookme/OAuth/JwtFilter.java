package com.hivestudent.bookme.OAuth;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.AllArgsConstructor;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.List;

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
        if (!jwtService.validateToken(token) || jwtService.isExpired(token)) {
            filterChain.doFilter(request, response);
            return;
        }
        // authentication
        var userSub = jwtService.extractEmail(token);
        var role = jwtService.extractRole(token);
        var authentication = new UsernamePasswordAuthenticationToken(
                userSub,
                null,
                List.of(new SimpleGrantedAuthority("ROLE_" + role))
        );

        //Attaching additional metadata about the request e.g ip address etc to the authentication object
        authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));

        //Stores current authentication to securityContextHolder, to access current user
        SecurityContextHolder.getContext().setAuthentication(authentication);

        //pass to next filter in chain
        filterChain.doFilter(request, response);
    }
}
