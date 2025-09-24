CREATE TABLE composers (
    composer_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    short_name VARCHAR(100) NOT NULL,
    full_name VARCHAR(200) NOT NULL
);

CREATE TABLE pieces (
    piece_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    composer_id INT NOT NULL REFERENCES composers(composer_id) ON DELETE CASCADE,
    piece_title VARCHAR(200) NOT NULL
);

CREATE TABLE programmes (
    programme_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    programme_title VARCHAR(200) NOT NULL
);

CREATE TABLE programme_pieces (
    programme_id INT REFERENCES programmes(programme_id) ON DELETE CASCADE,
    piece_id INT REFERENCES pieces(piece_id) ON DELETE CASCADE,
    sequence INT NOT NULL CHECK (sequence > 0),
    PRIMARY KEY (programme_id, piece_id),
    UNIQUE (programme_id, sequence)
);

CREATE TABLE venues (
    venue_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    address VARCHAR(100) NOT NULL
);

CREATE TABLE events (
    event_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_title VARCHAR(200) NOT NULL,
    venue_id INT REFERENCES venues(venue_id),
    event_date TIMESTAMP NOT NULL,
    ticket_link VARCHAR(500),
    programme_id INT NOT NULL REFERENCES programmes(programme_id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create a trigger for updating the created_at column
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach to the events table
CREATE TRIGGER update_events_updated_at
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
