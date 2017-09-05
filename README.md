# RSVPSs

This is a little server that allows you to display reservations on your static
website using the meetups.com API which is useful for sites like user groups or
other events like [meet-the-meetups.org](https://meet-the-meetups.org/) where we
have multiple events displayed on a single page.

Integrating this service consists of three big steps:

1. You have to install this service on a server and make it publicly
   accessible.
2. On the site you want to integrate RSVPs provide a simple text-file containing
   a list of all the meetups of which you want to fetch information. The RSVPs
   service will fetch this list to check that only those meetups can be fetched
   through the API that you want to.
3. Fetch the RSVPs from the service using a simple XHR (see "Sample
   JS-integration" for an example)


## Configuration

In order to configure the backend service, you have to pass a handful of
commandline arguments:

- `--verbose` enabled logging of debug messages
- `--meetup-api-key <VALUE>` allows you to specify and API key for interacting
  with the meetup.com-API. This is required in order to retrieve event
  information.
- `--whitelist-url <VALUE>` specifies the URL to a plaintext file that contains
  a list of valid meetup URLs. You can use this flag multiple times to specify
  multiple lists.
- `--whitelist-update-interval <DURATION>` defines the interval in which the
  whitelists should be requested from their respecive URLs and merged into a
  single big whitelist.
- `--http-addr <ADDR>` sets the HTTP address through which the service will be
  accessible via HTTP.
- `--allowed-origin <HOST>` defines a single host from which XHRs can be
  made. See
  [this article about CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) for
  details.
- `--cache-duration-upcoming <DURATION>` specifies the duration for which
  information of upcoming events should be kept in an internal cache.
- `--cache-duration-past <DURATION>` specifies the duration for which
  information of past events should be kept in an internal cache.


## Sample JS-integration

This is the code we initially used for [meet-the-meetups.org](https://github.com/zerok/meet-the-meetups.org/commit/da7a4e89c4fae50113a159f23e47b83e5b940845#diff-8a40c4132012f96d1558f15dda16fe38)?

```
(function() {
    var containers = Object.create(null);
    var urls = [];
    document.querySelectorAll('.event__group').forEach(function(event) {
    var url;
    var link = event.querySelector('.event__group__link');
    var placeholder = event.querySelector('.event__group__rsvps');
    if (link && link.getAttribute('href') && placeholder) {
        url = link.getAttribute('href');
        containers[url] = placeholder;
        urls.push(url);
    }
    });
    var req = new XMLHttpRequest();
    var reqData = JSON.stringify({meetups: urls});
    req.open('POST', 'https://api.meet-the-meetups.org/rsvps/query', true);
    req.addEventListener("load", function() {
    var data = JSON.parse(this.responseText);
    Object.keys(data).forEach(function(url) {
        var info = data[url];
        var container = containers[url];
        if (!container || !info || !info.rsvp) {
        return;
        }
        container.innerHTML = '(' + info.rsvp.YesCount + ' RSVPs of max ' + info.rsvp.MaxCount + ')';
    });
    });
    req.send(reqData);
})();
```
