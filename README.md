# Uni-Dock Server

## Usage
You should: 
- Make sure `unidock` and `fpocket` in your PATH.
- Or, set `UNIDOCK_PATH` and `FPOCKET_PATH` to your program location.

```sh
./unidock-server
```


## API

- Method: **Post**
- URL: `/unidock`
- Body:
    - File: receptor [.pdb,.pdbqt]
    - File: ligand [.pdb,.pdbqt]
- Response:
    - success:
        - Status: OK 200
        - Example body: `-7.552`
    - failed:
        - Status: InternalServerError 500
        - body: error message
    