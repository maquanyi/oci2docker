# Roadmap

This document defines the high-level goals of the oci2docker project. Its goal is to help both maintainers and contributors find meaningful tasks to focus on and create a low noise environment. The items in the 1.0 roadmap can be broken down into smaller milestones that are easy to accomplish.
    
## 1.0     

### Totally convert OCI bundle to Docker image and container

OCI bundle contains rootfs and runtime configurations. Converting OCI bundle not only mean just convert to docker image, but also extract runtime configurations and apply them to docker container.
       
### Supporting building Docker image 

Extracting all Docker image related elements(rootfs, items in configuration file) we can get from OCI bundle, then create Docker image based on them.
Using ocicovert to support testing of RKT   
        
### Supporing creating Docker container      
      
Based on container related configurations in OCI bundle, try to create Docker container.
