sub vcl_recv {
#FASTLY recv
  return(lookup);
}

sub vcl_hash {
  set req.hash += req.url;
  set req.hash += req.http.host;
#FASTLY hash
  return(hash);
}

sub vcl_hit {
#FASTLY hit
  return(deliver);
}

sub vcl_miss {
#FASTLY miss
  return(fetch);
}

sub vcl_pass {
#FASTLY pass
  return(pass);
}

sub vcl_fetch {
#FASTLY fetch
  return(deliver);
}

sub vcl_error {
#FASTLY error
  return(deliver);
}

sub vcl_deliver {
#FASTLY deliver
  set resp.http.X-Terraform-VCL-Test = "fixture";
  return(deliver);
}

sub vcl_log {
#FASTLY log
}
