# Uni-Dock Server
## Usage

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
    