import tkinter as tk
import time

# 144x speed: one full day (86400 sec) passes in 600 real seconds.
# That means each simulated second should advance every 600/86400 = 0.00694 sec (~7 ms)

UPDATE_INTERVAL = 7  # ms between each simulated second update

def update_clock():
    global sim_time

    # advance simulated time by 1 second
    sim_time += 1

    # wrap around 24h
    display = time.strftime("%H:%M:%S", time.gmtime(sim_time % 86400))
    label.config(text=display)

    # schedule next tick
    root.after(UPDATE_INTERVAL, update_clock)

# start at midnight
sim_time = 0

root = tk.Tk()
root.title("144x Speed Clock")

label = tk.Label(root, font=("Courier", 48), padx=20, pady=20)
label.pack()

update_clock()
root.mainloop()
