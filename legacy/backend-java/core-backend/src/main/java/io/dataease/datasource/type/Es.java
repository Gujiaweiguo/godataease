package io.dataease.datasource.type;


import org.springframework.stereotype.Component;

public class Es {
    private String url;
    private String username;
    private String password;
    private String version;
    private String uri;

    public String getUrl() { return url; }
    public void setUrl(String url) { this.url = url; }

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }

    public String getPassword() { return password; }
    public void setPassword(String password) { this.password = password; }

    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }

    public String getUri() { return uri; }
    public void setUri(String uri) { this.uri = uri; }
}
