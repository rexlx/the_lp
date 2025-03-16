from locust import HttpUser, task, between
import random
import string

class WebsiteUser(HttpUser):
    # Wait time between tasks for GET requests (2-6 seconds)
    wait_time = between(2, 6)
    
    # List of 10 predefined IDs
    id_list = [
        "user123",
        "test456",
        "client789",
        "abc101",
        "xyz202",
        "sample303",
        "data404",
        "info505",
        "query606",
        "load707"
    ]
    
    @task
    def get_access(self):
        # GET request to /access endpoint
        self.client.get("/query606")
    
    @task
    def post_tag(self):
        # Randomly select an ID from the list
        selected_id = random.choice(self.id_list)
        
        # Generate random hash
        random_hash = ''.join(random.choices(string.ascii_letters + string.digits, k=6))
        
        payload = {
            "id": selected_id,
            "hash": random_hash
        }
        
        # POST request to /tag endpoint
        self.client.post("/tag", json=payload)
        
        # Custom wait time for POST requests (5-15 seconds)
        from time import sleep
        from random import uniform
        sleep(uniform(5, 15))

# Optional: Add host if you want to specify the target server
# host = "http://localhost:8080"
