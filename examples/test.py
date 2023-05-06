import os


if __name__ == "__main__":
    # Example of emulating a data stream writing to a named pipe
    
    with open("pipe1", "w") as f:
        for i in range(20):
            f.write(f"LINE: {i}\n")
