# WebFendr

WebFendr is a bias, tiny web server that serves static sites with user login capabilities. 
It seamlessly integrates with storage buckets for quick deployment and is optimized 
for containerization, making it easy to manage in Kubernetes clusters.

The scenario WebFendr are meant for:

- You have a static site where you need to restrict access to all, or parts of the content.
- You already have a Kubernetes cluster (or a similar container image server)
- Your CDN/Storage provider does not let you add access control on top of your content.
- You already have a IAM solution setup for your company (eg: Auth0, Google IAM, Azure AD).

If you do not need access control on your site, WebFendr IS NOT for you. Here a traditional
CDN would make much more sense.

WebFendr can be installed as a single container, hosting all your static sites in one place. Or
it can be bundled into any container image, if you like to run one container per site. The best
strategy mainly depends on your build pipelines, and how your static sites are created, and deployed.

## How it works

WebFendr is a straightforward HTTP server, that serves static files from a local folder.
You can manually add these static files yourself or configure WebFendr to download them from a 
designated cloud storage location at regular intervals. You can add multiple sites as long as each 
site has its own domain that points to the server.
When WebFendr receives a request, it will first check the hostname in the request and verify 
whether the site exists and the user is logged in. If the requested resource requires user authentication, 
and the user is unauthorized, WebFendr will trigger the login flow. Once the  user has been authenticated, 
WebFendr will serve the content. So, no rocket science here.

NOTE: WebFendr has no SSL/HTTPS support at the moment, so it requires that you put an ingress in front of it, such as
Nginx or a Google LoadBalancer, that can serve certificates and terminate the SSL. We will consider adding
SSL support in the future. But it has not been prioritized since most Kubernetes clusters normally already will
have an ingress/load balancer installed.

## Installation

Installing WebFendr requires that you have a good understanding of modern tools and services, like Kubernetes,
Docker, Helm, and more. Depending of how you choose to install and run the service.

### Installing as a single container in Kubernetes

This is the default and recommended way of running WebFendr in production. 

#### With Helm


## Configuration

Coming soon...

## Resources

- https://auth0.com/docs/quickstart/webapp/golang/01-login

TODO

- Last-Modified on site files
- Cache-Control on site files
