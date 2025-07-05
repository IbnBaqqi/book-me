package com.hivestudent.bookme.Config;

import com.hivestudent.bookme.common.RateLimitFilter;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class FilterConfig {

    @Bean
    public FilterRegistrationBean<RateLimitFilter> rateLimitFilter() {
//        Manually register a Filter bean and set specific endpoints to target
//        Another way is to use MVC HandlerInterceptor(https://www.baeldung.com/spring-bucket4j#rect-api)
        FilterRegistrationBean<RateLimitFilter> registrationBean = new FilterRegistrationBean<>();

        RateLimitFilter filter = new RateLimitFilter();
        registrationBean.setFilter(filter);

        registrationBean.addUrlPatterns("/reservations/*");
        registrationBean.setOrder(1);

        return registrationBean;
    }
}
