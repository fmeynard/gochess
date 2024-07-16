#!/bin/bash

# Initialize the path to Stockfish as an empty string, to use the default path
stockfish_path="stockfish"  # Assume stockfish is in the PATH
custom_engine_path="cmd/perft.go"

# Check for the --stockfish flag and set custom path if provided
for arg in "$@"; do
  case $arg in
    --stockfish=*)
      stockfish_path="${arg#*=}"
      shift # Remove the --stockfish argument
      ;;
    --custom-path=*)
      custom_engine_path="${arg#*=}"
      shift # Remove the --stockfish argument
      ;;
    *)
      # Collect the other arguments
      positional_args+=("$arg")
      ;;
  esac
done

# Ensure the correct number of positional arguments are passed
if [ "${#positional_args[@]}" -ne 2 ]; then
  echo "Usage: $0 fen depth [--stockfish=path_to_stockfish] [--custom-path=path_to_engine.go]"
  exit 1
fi

fen="${positional_args[0]}"
depth="${positional_args[1]}"

if [ ! -f "$custom_engine_path" ]; then
  echo "Custom engine not found at $custom_engine_path."
  exit 1
fi

# Get perft results from custom engine
custom_perft_file=$(mktemp)
custom_engine_command="go run $custom_engine_path \"$fen\" $depth"
eval $custom_engine_command | grep -Eo '^[a-h][1-8][a-h][1-8]: \d+' | sort > "$custom_perft_file"

# Get perft results from Stockfish
stockfish_perft_file=$(mktemp)

$stockfish_path << EOF | grep -Eo '^[a-h][1-8][a-h][1-8]: \d+' | sort > "$stockfish_perft_file"
uci
position fen $fen
go perft $depth
quit
EOF

# Arrays to store position identifiers and moves count
custom_positions=()
custom_moves=()
stockfish_positions=()
stockfish_moves=()

# Function to read file into arrays
read_file_into_arrays() {
  local file=$1
  local positions=()
  local moves=()

  while IFS=" " read -r pos moves_count; do
    pos=${pos%:} # Remove trailing colon
    positions+=("$pos")
    moves+=("$moves_count")
  done < "$file"

  # Return arrays
  echo "${positions[@]}"
  echo "${moves[@]}"
}

# Read data from custom engine and Stockfish results into arrays
custom_data=($(read_file_into_arrays "$custom_perft_file"))
custom_positions=(${custom_data[@]:0:${#custom_data[@]}/2})
custom_moves=(${custom_data[@]:${#custom_data[@]}/2})

stockfish_data=($(read_file_into_arrays "$stockfish_perft_file"))
stockfish_positions=(${stockfish_data[@]:0:${#stockfish_data[@]}/2})
stockfish_moves=(${stockfish_data[@]:${#stockfish_data[@]}/2})

# Get all unique position identifiers
all_positions=$(printf "%s\n" "${custom_positions[@]}" "${stockfish_positions[@]}" | sort | uniq)

# Print the table header
echo -e "Position\tCustom Engine Moves\tStockfish Moves\tMatch"

# Print the table rows
for pos in $all_positions; do
  moves1="MISSING"
  moves2="MISSING"
  for ((i = 0; i < ${#custom_positions[@]}; i++)); do
    if [ "${custom_positions[i]}" = "$pos" ]; then
      moves1="${custom_moves[i]}"
      break
    fi
  done
  for ((i = 0; i < ${#stockfish_positions[@]}; i++)); do
    if [ "${stockfish_positions[i]}" = "$pos" ]; then
      moves2="${stockfish_moves[i]}"
      break
    fi
  done
  match="<<< KO"
  if [ "$moves2" = "$moves1" ]; then
    match=""
  fi
  echo -e "$pos\t$moves1\t$moves2\t$match"
done

# Clean up temporary files
rm "$custom_perft_file" "$stockfish_perft_file"