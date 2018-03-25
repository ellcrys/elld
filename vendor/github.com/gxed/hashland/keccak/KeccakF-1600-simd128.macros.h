/*
The Keccak sponge function, designed by Guido Bertoni, Joan Daemen,
Michaël Peeters and Gilles Van Assche. For more information, feedback or
questions, please refer to our website: http://keccak.noekeon.org/

Implementation by the designers,
hereby denoted as "the implementer".

To the extent possible under law, the implementer has waived all copyright
and related or neighboring rights to the source code in this file.
http://creativecommons.org/publicdomain/zero/1.0/
*/

#define declareABCDE \
    V6464 Abage, Abegi, Abigo, Abogu, Abuga; \
    V6464 Akame, Akemi, Akimo, Akomu, Akuma; \
    V6464 Abae, Abio, Agae, Agio, Akae, Akio, Amae, Amio, Asae, Asio; \
    V64 Aba, Abe, Abi, Abo, Abu; \
    V64 Aga, Age, Agi, Ago, Agu; \
    V64 Aka, Ake, Aki, Ako, Aku; \
    V64 Ama, Ame, Ami, Amo, Amu; \
    V64 Asa, Ase, Asi, Aso, Asu; \
    V128 Bbage, Bbegi, Bbigo, Bbogu, Bbuga; \
    V128 Bkame, Bkemi, Bkimo, Bkomu, Bkuma; \
    V64 Bba, Bbe, Bbi, Bbo, Bbu; \
    V64 Bga, Bge, Bgi, Bgo, Bgu; \
    V64 Bka, Bke, Bki, Bko, Bku; \
    V64 Bma, Bme, Bmi, Bmo, Bmu; \
    V64 Bsa, Bse, Bsi, Bso, Bsu; \
    V128 Cae, Cei, Cio, Cou, Cua, Dei, Dou; \
    V64 Ca, Ce, Ci, Co, Cu; \
    V64 Da, De, Di, Do, Du; \
    V6464 Ebage, Ebegi, Ebigo, Ebogu, Ebuga; \
    V6464 Ekame, Ekemi, Ekimo, Ekomu, Ekuma; \
    V64 Eba, Ebe, Ebi, Ebo, Ebu; \
    V64 Ega, Ege, Egi, Ego, Egu; \
    V64 Eka, Eke, Eki, Eko, Eku; \
    V64 Ema, Eme, Emi, Emo, Emu; \
    V64 Esa, Ese, Esi, Eso, Esu; \
    V128 Zero;

#define prepareTheta

#define computeD \
    Cua = GET64LO(Cu, Cae); \
    Dei = XOR128(Cae, ROL64in128(Cio, 1)); \
    Dou = XOR128(Cio, ROL64in128(Cua, 1)); \
    Da = XOR64(Cu, ROL64in128(COPY64HI2LO(Cae), 1)); \
    De = Dei; \
    Di = COPY64HI2LO(Dei); \
    Do = Dou; \
    Du = COPY64HI2LO(Dou);

// --- Theta Rho Pi Chi Iota Prepare-theta
// --- 64-bit lanes mapped to 64-bit and 128-bit words
#define thetaRhoPiChiIotaPrepareTheta(i, A, E) \
    computeD \
    \
    A##ba = LOAD64(A##bage.v64[0]); \
    XOReq64(A##ba, Da); \
    Bba = A##ba; \
    XOReq64(A##gu, Du); \
    Bge = ROL64(A##gu, 20); \
    Bbage = GET64LO(Bba, Bge); \
    A##ge = LOAD64(A##bage.v64[1]); \
    XOReq64(A##ge, De); \
    Bbe = ROL64(A##ge, 44); \
    A##ka = LOAD64(A##kame.v64[0]); \
    XOReq64(A##ka, Da); \
    Bgi = ROL64(A##ka, 3); \
    Bbegi = GET64LO(Bbe, Bgi); \
    XOReq64(A##ki, Di); \
    Bbi = ROL64(A##ki, 43); \
    A##me = LOAD64(A##kame.v64[1]); \
    XOReq64(A##me, De); \
    Bgo = ROL64(A##me, 45); \
    Bbigo = GET64LO(Bbi, Bgo); \
    E##bage.v128 = XOR128(Bbage, ANDnu128(Bbegi, Bbigo)); \
    XOReq128(E##bage.v128, CONST64(KeccakF1600RoundConstants[i])); \
    Cae = E##bage.v128; \
    XOReq64(A##mo, Do); \
    Bbo = ROL64(A##mo, 21); \
    XOReq64(A##si, Di); \
    Bgu = ROL64(A##si, 61); \
    Bbogu = GET64LO(Bbo, Bgu); \
    E##begi.v128 = XOR128(Bbegi, ANDnu128(Bbigo, Bbogu)); \
    Cei = E##begi.v128; \
    XOReq64(A##su, Du); \
    Bbu = ROL64(A##su, 14); \
    XOReq64(A##bo, Do); \
    Bga = ROL64(A##bo, 28); \
    Bbuga = GET64LO(Bbu, Bga); \
    E##bigo.v128 = XOR128(Bbigo, ANDnu128(Bbogu, Bbuga)); \
    E##bi = E##bigo.v128; \
    E##go = GET64HI(E##bigo.v128, E##bigo.v128); \
    Cio = E##bigo.v128; \
    E##bogu.v128 = XOR128(Bbogu, ANDnu128(Bbuga, Bbage)); \
    E##bo = E##bogu.v128; \
    E##gu = GET64HI(E##bogu.v128, E##bogu.v128); \
    Cou = E##bogu.v128; \
    E##buga.v128 = XOR128(Bbuga, ANDnu128(Bbage, Bbegi)); \
    E##bu = E##buga.v128; \
    E##ga = GET64HI(E##buga.v128, E##buga.v128); \
    Cua = E##buga.v128; \
\
    A##be = LOAD64(A##begi.v64[0]); \
    XOReq64(A##be, De); \
    Bka = ROL64(A##be, 1); \
    XOReq64(A##ga, Da); \
    Bme = ROL64(A##ga, 36); \
    Bkame = GET64LO(Bka, Bme); \
    A##gi = LOAD64(A##begi.v64[1]); \
    XOReq64(A##gi, Di); \
    Bke = ROL64(A##gi, 6); \
    A##ke = LOAD64(A##kemi.v64[0]); \
    XOReq64(A##ke, De); \
    Bmi = ROL64(A##ke, 10); \
    Bkemi = GET64LO(Bke, Bmi); \
    XOReq64(A##ko, Do); \
    Bki = ROL64(A##ko, 25); \
    A##mi = LOAD64(A##kemi.v64[1]); \
    XOReq64(A##mi, Di); \
    Bmo = ROL64(A##mi, 15); \
    Bkimo = GET64LO(Bki, Bmo); \
    E##kame.v128 = XOR128(Bkame, ANDnu128(Bkemi, Bkimo)); \
    XOReq128(Cae, E##kame.v128); \
    XOReq64(A##mu, Du); \
    Bko = ROL64(A##mu, 8); \
    XOReq64(A##so, Do); \
    Bmu = ROL64(A##so, 56); \
    Bkomu = GET64LO(Bko, Bmu); \
    E##kemi.v128 = XOR128(Bkemi, ANDnu128(Bkimo, Bkomu)); \
    XOReq128(Cei, E##kemi.v128); \
    XOReq64(A##sa, Da); \
    Bku = ROL64(A##sa, 18); \
    XOReq64(A##bu, Du); \
    Bma = ROL64(A##bu, 27); \
    Bkuma = GET64LO(Bku, Bma); \
    E##kimo.v128 = XOR128(Bkimo, ANDnu128(Bkomu, Bkuma)); \
    E##ki = E##kimo.v128; \
    E##mo = GET64HI(E##kimo.v128, E##kimo.v128); \
    XOReq128(Cio, E##kimo.v128); \
    E##komu.v128 = XOR128(Bkomu, ANDnu128(Bkuma, Bkame)); \
    E##ko = E##komu.v128; \
    E##mu = GET64HI(E##komu.v128, E##komu.v128); \
    XOReq128(Cou, E##komu.v128); \
    E##kuma.v128 = XOR128(Bkuma, ANDnu128(Bkame, Bkemi)); \
    E##ku = E##kuma.v128; \
    E##ma = GET64HI(E##kuma.v128, E##kuma.v128); \
    XOReq128(Cua, E##kuma.v128); \
\
    XOReq64(A##bi, Di); \
    Bsa = ROL64(A##bi, 62); \
    XOReq64(A##go, Do); \
    Bse = ROL64(A##go, 55); \
    XOReq64(A##ku, Du); \
    Bsi = ROL64(A##ku, 39); \
    E##sa = XOR64(Bsa, ANDnu64(Bse, Bsi)); \
    Ca = E##sa; \
    XOReq64(A##ma, Da); \
    Bso = ROL64(A##ma, 41); \
    E##se = XOR64(Bse, ANDnu64(Bsi, Bso)); \
    Ce = E##se; \
    XOReq128(Cae, GET64LO(Ca, Ce)); \
    XOReq64(A##se, De); \
    Bsu = ROL64(A##se, 2); \
    E##si = XOR64(Bsi, ANDnu64(Bso, Bsu)); \
    Ci = E##si; \
    E##so = XOR64(Bso, ANDnu64(Bsu, Bsa)); \
    Co = E##so; \
    XOReq128(Cio, GET64LO(Ci, Co)); \
    E##su = XOR64(Bsu, ANDnu64(Bsa, Bse)); \
    Cu = E##su; \
\
    Zero = ZERO128(); \
    XOReq128(Cae, GET64HI(Cua, Zero)); \
    XOReq128(Cae, GET64LO(Zero, Cei)); \
    XOReq128(Cio, GET64HI(Cei, Zero)); \
    XOReq128(Cio, GET64LO(Zero, Cou)); \
    XOReq128(Cua, GET64HI(Cou, Zero)); \
    XOReq64(Cu, Cua); \

// --- Theta Rho Pi Chi Iota
// --- 64-bit lanes mapped to 64-bit and 128-bit words
#define thetaRhoPiChiIota(i, A, E) thetaRhoPiChiIotaPrepareTheta(i, A, E)

const UINT64 KeccakF1600RoundConstants[24] = {
    0x0000000000000001ULL,
    0x0000000000008082ULL,
    0x800000000000808aULL,
    0x8000000080008000ULL,
    0x000000000000808bULL,
    0x0000000080000001ULL,
    0x8000000080008081ULL,
    0x8000000000008009ULL,
    0x000000000000008aULL,
    0x0000000000000088ULL,
    0x0000000080008009ULL,
    0x000000008000000aULL,
    0x000000008000808bULL,
    0x800000000000008bULL,
    0x8000000000008089ULL,
    0x8000000000008003ULL,
    0x8000000000008002ULL,
    0x8000000000000080ULL,
    0x000000000000800aULL,
    0x800000008000000aULL,
    0x8000000080008081ULL,
    0x8000000000008080ULL,
    0x0000000080000001ULL,
    0x8000000080008008ULL };

#define copyFromStateAndXor576bits(X, state, input) \
    X##bae.v128 = XOR128(LOAD128(state[ 0]), LOAD128u(input[ 0])); \
    X##ba = X##bae.v128; \
    X##be = GET64HI(X##bae.v128, X##bae.v128); \
    Cae = X##bae.v128; \
    X##bio.v128 = XOR128(LOAD128(state[ 2]), LOAD128u(input[ 2])); \
    X##bi = X##bio.v128; \
    X##bo = GET64HI(X##bio.v128, X##bio.v128); \
    Cio = X##bio.v128; \
    X##bu = XOR64(LOAD64(state[ 4]), LOAD64(input[ 4])); \
    Cu = X##bu; \
    X##gae.v128 = XOR128(LOAD128u(state[ 5]), LOAD128u(input[ 5])); \
    X##ga = X##gae.v128; \
    X##ge = GET64HI(X##gae.v128, X##gae.v128); \
    X##bage.v128 = GET64LO(X##ba, X##ge); \
    XOReq128(Cae, X##gae.v128); \
    X##gio.v128 = XOR128(LOAD128u(state[ 7]), LOAD128u(input[ 7])); \
    X##gi = X##gio.v128; \
    X##begi.v128 = GET64LO(X##be, X##gi); \
    X##go = GET64HI(X##gio.v128, X##gio.v128); \
    XOReq128(Cio, X##gio.v128); \
    X##gu = LOAD64(state[ 9]); \
    XOReq64(Cu, X##gu); \
    X##kae.v128 = LOAD128(state[10]); \
    X##ka = X##kae.v128; \
    X##ke = GET64HI(X##kae.v128, X##kae.v128); \
    XOReq128(Cae, X##kae.v128); \
    X##kio.v128 = LOAD128(state[12]); \
    X##ki = X##kio.v128; \
    X##ko = GET64HI(X##kio.v128, X##kio.v128); \
    XOReq128(Cio, X##kio.v128); \
    X##ku = LOAD64(state[14]); \
    XOReq64(Cu, X##ku); \
    X##mae.v128 = LOAD128u(state[15]); \
    X##ma = X##mae.v128; \
    X##me = GET64HI(X##mae.v128, X##mae.v128); \
    X##kame.v128 = GET64LO(X##ka, X##me); \
    XOReq128(Cae, X##mae.v128); \
    X##mio.v128 = LOAD128u(state[17]); \
    X##mi = X##mio.v128; \
    X##kemi.v128 = GET64LO(X##ke, X##mi); \
    X##mo = GET64HI(X##mio.v128, X##mio.v128); \
    XOReq128(Cio, X##mio.v128); \
    X##mu = LOAD64(state[19]); \
    XOReq64(Cu, X##mu); \
    X##sae.v128 = LOAD128(state[20]); \
    X##sa = X##sae.v128; \
    X##se = GET64HI(X##sae.v128, X##sae.v128); \
    XOReq128(Cae, X##sae.v128); \
    X##sio.v128 = LOAD128(state[22]); \
    X##si = X##sio.v128; \
    X##so = GET64HI(X##sio.v128, X##sio.v128); \
    XOReq128(Cio, X##sio.v128); \
    X##su = LOAD64(state[24]); \
    XOReq64(Cu, X##su); \

#define copyFromStateAndXor832bits(X, state, input) \
    X##bae.v128 = XOR128(LOAD128(state[ 0]), LOAD128u(input[ 0])); \
    X##ba = X##bae.v128; \
    X##be = GET64HI(X##bae.v128, X##bae.v128); \
    Cae = X##bae.v128; \
    X##bio.v128 = XOR128(LOAD128(state[ 2]), LOAD128u(input[ 2])); \
    X##bi = X##bio.v128; \
    X##bo = GET64HI(X##bio.v128, X##bio.v128); \
    Cio = X##bio.v128; \
    X##bu = XOR64(LOAD64(state[ 4]), LOAD64(input[ 4])); \
    Cu = X##bu; \
    X##gae.v128 = XOR128(LOAD128u(state[ 5]), LOAD128u(input[ 5])); \
    X##ga = X##gae.v128; \
    X##ge = GET64HI(X##gae.v128, X##gae.v128); \
    X##bage.v128 = GET64LO(X##ba, X##ge); \
    XOReq128(Cae, X##gae.v128); \
    X##gio.v128 = XOR128(LOAD128u(state[ 7]), LOAD128u(input[ 7])); \
    X##gi = X##gio.v128; \
    X##begi.v128 = GET64LO(X##be, X##gi); \
    X##go = GET64HI(X##gio.v128, X##gio.v128); \
    XOReq128(Cio, X##gio.v128); \
    X##gu = XOR64(LOAD64(state[ 9]), LOAD64(input[ 9])); \
    XOReq64(Cu, X##gu); \
    X##kae.v128 = XOR128(LOAD128(state[10]), LOAD128u(input[10])); \
    X##ka = X##kae.v128; \
    X##ke = GET64HI(X##kae.v128, X##kae.v128); \
    XOReq128(Cae, X##kae.v128); \
    X##kio.v128 = XOR128(LOAD128(state[12]), LOAD64(input[12])); \
    X##ki = X##kio.v128; \
    X##ko = GET64HI(X##kio.v128, X##kio.v128); \
    XOReq128(Cio, X##kio.v128); \
    X##ku = LOAD64(state[14]); \
    XOReq64(Cu, X##ku); \
    X##mae.v128 = LOAD128u(state[15]); \
    X##ma = X##mae.v128; \
    X##me = GET64HI(X##mae.v128, X##mae.v128); \
    X##kame.v128 = GET64LO(X##ka, X##me); \
    XOReq128(Cae, X##mae.v128); \
    X##mio.v128 = LOAD128u(state[17]); \
    X##mi = X##mio.v128; \
    X##kemi.v128 = GET64LO(X##ke, X##mi); \
    X##mo = GET64HI(X##mio.v128, X##mio.v128); \
    XOReq128(Cio, X##mio.v128); \
    X##mu = LOAD64(state[19]); \
    XOReq64(Cu, X##mu); \
    X##sae.v128 = LOAD128(state[20]); \
    X##sa = X##sae.v128; \
    X##se = GET64HI(X##sae.v128, X##sae.v128); \
    XOReq128(Cae, X##sae.v128); \
    X##sio.v128 = LOAD128(state[22]); \
    X##si = X##sio.v128; \
    X##so = GET64HI(X##sio.v128, X##sio.v128); \
    XOReq128(Cio, X##sio.v128); \
    X##su = LOAD64(state[24]); \
    XOReq64(Cu, X##su); \

#define copyFromStateAndXor1024bits(X, state, input) \
    X##bae.v128 = XOR128(LOAD128(state[ 0]), LOAD128u(input[ 0])); \
    X##ba = X##bae.v128; \
    X##be = GET64HI(X##bae.v128, X##bae.v128); \
    Cae = X##bae.v128; \
    X##bio.v128 = XOR128(LOAD128(state[ 2]), LOAD128u(input[ 2])); \
    X##bi = X##bio.v128; \
    X##bo = GET64HI(X##bio.v128, X##bio.v128); \
    Cio = X##bio.v128; \
    X##bu = XOR64(LOAD64(state[ 4]), LOAD64(input[ 4])); \
    Cu = X##bu; \
    X##gae.v128 = XOR128(LOAD128u(state[ 5]), LOAD128u(input[ 5])); \
    X##ga = X##gae.v128; \
    X##ge = GET64HI(X##gae.v128, X##gae.v128); \
    X##bage.v128 = GET64LO(X##ba, X##ge); \
    XOReq128(Cae, X##gae.v128); \
    X##gio.v128 = XOR128(LOAD128u(state[ 7]), LOAD128u(input[ 7])); \
    X##gi = X##gio.v128; \
    X##begi.v128 = GET64LO(X##be, X##gi); \
    X##go = GET64HI(X##gio.v128, X##gio.v128); \
    XOReq128(Cio, X##gio.v128); \
    X##gu = XOR64(LOAD64(state[ 9]), LOAD64(input[ 9])); \
    XOReq64(Cu, X##gu); \
    X##kae.v128 = XOR128(LOAD128(state[10]), LOAD128u(input[10])); \
    X##ka = X##kae.v128; \
    X##ke = GET64HI(X##kae.v128, X##kae.v128); \
    XOReq128(Cae, X##kae.v128); \
    X##kio.v128 = XOR128(LOAD128(state[12]), LOAD128u(input[12])); \
    X##ki = X##kio.v128; \
    X##ko = GET64HI(X##kio.v128, X##kio.v128); \
    XOReq128(Cio, X##kio.v128); \
    X##ku = XOR64(LOAD64(state[14]), LOAD64(input[14])); \
    XOReq64(Cu, X##ku); \
    X##mae.v128 = XOR128(LOAD128u(state[15]), LOAD64(input[15])); \
    X##ma = X##mae.v128; \
    X##me = GET64HI(X##mae.v128, X##mae.v128); \
    X##kame.v128 = GET64LO(X##ka, X##me); \
    XOReq128(Cae, X##mae.v128); \
    X##mio.v128 = LOAD128u(state[17]); \
    X##mi = X##mio.v128; \
    X##kemi.v128 = GET64LO(X##ke, X##mi); \
    X##mo = GET64HI(X##mio.v128, X##mio.v128); \
    XOReq128(Cio, X##mio.v128); \
    X##mu = LOAD64(state[19]); \
    XOReq64(Cu, X##mu); \
    X##sae.v128 = LOAD128(state[20]); \
    X##sa = X##sae.v128; \
    X##se = GET64HI(X##sae.v128, X##sae.v128); \
    XOReq128(Cae, X##sae.v128); \
    X##sio.v128 = LOAD128(state[22]); \
    X##si = X##sio.v128; \
    X##so = GET64HI(X##sio.v128, X##sio.v128); \
    XOReq128(Cio, X##sio.v128); \
    X##su = LOAD64(state[24]); \
    XOReq64(Cu, X##su); \

#define copyFromStateAndXor1088bits(X, state, input) \
    X##bae.v128 = XOR128(LOAD128(state[ 0]), LOAD128u(input[ 0])); \
    X##ba = X##bae.v128; \
    X##be = GET64HI(X##bae.v128, X##bae.v128); \
    Cae = X##bae.v128; \
    X##bio.v128 = XOR128(LOAD128(state[ 2]), LOAD128u(input[ 2])); \
    X##bi = X##bio.v128; \
    X##bo = GET64HI(X##bio.v128, X##bio.v128); \
    Cio = X##bio.v128; \
    X##bu = XOR64(LOAD64(state[ 4]), LOAD64(input[ 4])); \
    Cu = X##bu; \
    X##gae.v128 = XOR128(LOAD128u(state[ 5]), LOAD128u(input[ 5])); \
    X##ga = X##gae.v128; \
    X##ge = GET64HI(X##gae.v128, X##gae.v128); \
    X##bage.v128 = GET64LO(X##ba, X##ge); \
    XOReq128(Cae, X##gae.v128); \
    X##gio.v128 = XOR128(LOAD128u(state[ 7]), LOAD128u(input[ 7])); \
    X##gi = X##gio.v128; \
    X##begi.v128 = GET64LO(X##be, X##gi); \
    X##go = GET64HI(X##gio.v128, X##gio.v128); \
    XOReq128(Cio, X##gio.v128); \
    X##gu = XOR64(LOAD64(state[ 9]), LOAD64(input[ 9])); \
    XOReq64(Cu, X##gu); \
    X##kae.v128 = XOR128(LOAD128(state[10]), LOAD128u(input[10])); \
    X##ka = X##kae.v128; \
    X##ke = GET64HI(X##kae.v128, X##kae.v128); \
    XOReq128(Cae, X##kae.v128); \
    X##kio.v128 = XOR128(LOAD128(state[12]), LOAD128u(input[12])); \
    X##ki = X##kio.v128; \
    X##ko = GET64HI(X##kio.v128, X##kio.v128); \
    XOReq128(Cio, X##kio.v128); \
    X##ku = XOR64(LOAD64(state[14]), LOAD64(input[14])); \
    XOReq64(Cu, X##ku); \
    X##mae.v128 = XOR128(LOAD128u(state[15]), LOAD128u(input[15])); \
    X##ma = X##mae.v128; \
    X##me = GET64HI(X##mae.v128, X##mae.v128); \
    X##kame.v128 = GET64LO(X##ka, X##me); \
    XOReq128(Cae, X##mae.v128); \
    X##mio.v128 = LOAD128u(state[17]); \
    X##mi = X##mio.v128; \
    X##kemi.v128 = GET64LO(X##ke, X##mi); \
    X##mo = GET64HI(X##mio.v128, X##mio.v128); \
    XOReq128(Cio, X##mio.v128); \
    X##mu = LOAD64(state[19]); \
    XOReq64(Cu, X##mu); \
    X##sae.v128 = LOAD128(state[20]); \
    X##sa = X##sae.v128; \
    X##se = GET64HI(X##sae.v128, X##sae.v128); \
    XOReq128(Cae, X##sae.v128); \
    X##sio.v128 = LOAD128(state[22]); \
    X##si = X##sio.v128; \
    X##so = GET64HI(X##sio.v128, X##sio.v128); \
    XOReq128(Cio, X##sio.v128); \
    X##su = LOAD64(state[24]); \
    XOReq64(Cu, X##su); \

#define copyFromStateAndXor1152bits(X, state, input) \
    X##bae.v128 = XOR128(LOAD128(state[ 0]), LOAD128u(input[ 0])); \
    X##ba = X##bae.v128; \
    X##be = GET64HI(X##bae.v128, X##bae.v128); \
    Cae = X##bae.v128; \
    X##bio.v128 = XOR128(LOAD128(state[ 2]), LOAD128u(input[ 2])); \
    X##bi = X##bio.v128; \
    X##bo = GET64HI(X##bio.v128, X##bio.v128); \
    Cio = X##bio.v128; \
    X##bu = XOR64(LOAD64(state[ 4]), LOAD64(input[ 4])); \
    Cu = X##bu; \
    X##gae.v128 = XOR128(LOAD128u(state[ 5]), LOAD128u(input[ 5])); \
    X##ga = X##gae.v128; \
    X##ge = GET64HI(X##gae.v128, X##gae.v128); \
    X##bage.v128 = GET64LO(X##ba, X##ge); \
    XOReq128(Cae, X##gae.v128); \
    X##gio.v128 = XOR128(LOAD128u(state[ 7]), LOAD128u(input[ 7])); \
    X##gi = X##gio.v128; \
    X##begi.v128 = GET64LO(X##be, X##gi); \
    X##go = GET64HI(X##gio.v128, X##gio.v128); \
    XOReq128(Cio, X##gio.v128); \
    X##gu = XOR64(LOAD64(state[ 9]), LOAD64(input[ 9])); \
    XOReq64(Cu, X##gu); \
    X##kae.v128 = XOR128(LOAD128(state[10]), LOAD128u(input[10])); \
    X##ka = X##kae.v128; \
    X##ke = GET64HI(X##kae.v128, X##kae.v128); \
    XOReq128(Cae, X##kae.v128); \
    X##kio.v128 = XOR128(LOAD128(state[12]), LOAD128u(input[12])); \
    X##ki = X##kio.v128; \
    X##ko = GET64HI(X##kio.v128, X##kio.v128); \
    XOReq128(Cio, X##kio.v128); \
    X##ku = XOR64(LOAD64(state[14]), LOAD64(input[14])); \
    XOReq64(Cu, X##ku); \
    X##mae.v128 = XOR128(LOAD128u(state[15]), LOAD128u(input[15])); \
    X##ma = X##mae.v128; \
    X##me = GET64HI(X##mae.v128, X##mae.v128); \
    X##kame.v128 = GET64LO(X##ka, X##me); \
    XOReq128(Cae, X##mae.v128); \
    X##mio.v128 = XOR128(LOAD128u(state[17]), LOAD64(input[17])); \
    X##mi = X##mio.v128; \
    X##kemi.v128 = GET64LO(X##ke, X##mi); \
    X##mo = GET64HI(X##mio.v128, X##mio.v128); \
    XOReq128(Cio, X##mio.v128); \
    X##mu = LOAD64(state[19]); \
    XOReq64(Cu, X##mu); \
    X##sae.v128 = LOAD128(state[20]); \
    X##sa = X##sae.v128; \
    X##se = GET64HI(X##sae.v128, X##sae.v128); \
    XOReq128(Cae, X##sae.v128); \
    X##sio.v128 = LOAD128(state[22]); \
    X##si = X##sio.v128; \
    X##so = GET64HI(X##sio.v128, X##sio.v128); \
    XOReq128(Cio, X##sio.v128); \
    X##su = LOAD64(state[24]); \
    XOReq64(Cu, X##su); \

#define copyFromStateAndXor1344bits(X, state, input) \
    X##bae.v128 = XOR128(LOAD128(state[ 0]), LOAD128u(input[ 0])); \
    X##ba = X##bae.v128; \
    X##be = GET64HI(X##bae.v128, X##bae.v128); \
    Cae = X##bae.v128; \
    X##bio.v128 = XOR128(LOAD128(state[ 2]), LOAD128u(input[ 2])); \
    X##bi = X##bio.v128; \
    X##bo = GET64HI(X##bio.v128, X##bio.v128); \
    Cio = X##bio.v128; \
    X##bu = XOR64(LOAD64(state[ 4]), LOAD64(input[ 4])); \
    Cu = X##bu; \
    X##gae.v128 = XOR128(LOAD128u(state[ 5]), LOAD128u(input[ 5])); \
    X##ga = X##gae.v128; \
    X##ge = GET64HI(X##gae.v128, X##gae.v128); \
    X##bage.v128 = GET64LO(X##ba, X##ge); \
    XOReq128(Cae, X##gae.v128); \
    X##gio.v128 = XOR128(LOAD128u(state[ 7]), LOAD128u(input[ 7])); \
    X##gi = X##gio.v128; \
    X##begi.v128 = GET64LO(X##be, X##gi); \
    X##go = GET64HI(X##gio.v128, X##gio.v128); \
    XOReq128(Cio, X##gio.v128); \
    X##gu = XOR64(LOAD64(state[ 9]), LOAD64(input[ 9])); \
    XOReq64(Cu, X##gu); \
    X##kae.v128 = XOR128(LOAD128(state[10]), LOAD128u(input[10])); \
    X##ka = X##kae.v128; \
    X##ke = GET64HI(X##kae.v128, X##kae.v128); \
    XOReq128(Cae, X##kae.v128); \
    X##kio.v128 = XOR128(LOAD128(state[12]), LOAD128u(input[12])); \
    X##ki = X##kio.v128; \
    X##ko = GET64HI(X##kio.v128, X##kio.v128); \
    XOReq128(Cio, X##kio.v128); \
    X##ku = XOR64(LOAD64(state[14]), LOAD64(input[14])); \
    XOReq64(Cu, X##ku); \
    X##mae.v128 = XOR128(LOAD128u(state[15]), LOAD128u(input[15])); \
    X##ma = X##mae.v128; \
    X##me = GET64HI(X##mae.v128, X##mae.v128); \
    X##kame.v128 = GET64LO(X##ka, X##me); \
    XOReq128(Cae, X##mae.v128); \
    X##mio.v128 = XOR128(LOAD128u(state[17]), LOAD128u(input[17])); \
    X##mi = X##mio.v128; \
    X##kemi.v128 = GET64LO(X##ke, X##mi); \
    X##mo = GET64HI(X##mio.v128, X##mio.v128); \
    XOReq128(Cio, X##mio.v128); \
    X##mu = XOR64(LOAD64(state[19]), LOAD64(input[19])); \
    XOReq64(Cu, X##mu); \
    X##sae.v128 = XOR128(LOAD128(state[20]), LOAD64(input[20])); \
    X##sa = X##sae.v128; \
    X##se = GET64HI(X##sae.v128, X##sae.v128); \
    XOReq128(Cae, X##sae.v128); \
    X##sio.v128 = LOAD128(state[22]); \
    X##si = X##sio.v128; \
    X##so = GET64HI(X##sio.v128, X##sio.v128); \
    XOReq128(Cio, X##sio.v128); \
    X##su = LOAD64(state[24]); \
    XOReq64(Cu, X##su); \

#define copyFromState(X, state) \
    X##bae.v128 = LOAD128(state[ 0]); \
    X##ba = X##bae.v128; \
    X##be = GET64HI(X##bae.v128, X##bae.v128); \
    Cae = X##bae.v128; \
    X##bio.v128 = LOAD128(state[ 2]); \
    X##bi = X##bio.v128; \
    X##bo = GET64HI(X##bio.v128, X##bio.v128); \
    Cio = X##bio.v128; \
    X##bu = LOAD64(state[ 4]); \
    Cu = X##bu; \
    X##gae.v128 = LOAD128u(state[ 5]); \
    X##ga = X##gae.v128; \
    X##ge = GET64HI(X##gae.v128, X##gae.v128); \
    X##bage.v128 = GET64LO(X##ba, X##ge); \
    XOReq128(Cae, X##gae.v128); \
    X##gio.v128 = LOAD128u(state[ 7]); \
    X##gi = X##gio.v128; \
    X##begi.v128 = GET64LO(X##be, X##gi); \
    X##go = GET64HI(X##gio.v128, X##gio.v128); \
    XOReq128(Cio, X##gio.v128); \
    X##gu = LOAD64(state[ 9]); \
    XOReq64(Cu, X##gu); \
    X##kae.v128 = LOAD128(state[10]); \
    X##ka = X##kae.v128; \
    X##ke = GET64HI(X##kae.v128, X##kae.v128); \
    XOReq128(Cae, X##kae.v128); \
    X##kio.v128 = LOAD128(state[12]); \
    X##ki = X##kio.v128; \
    X##ko = GET64HI(X##kio.v128, X##kio.v128); \
    XOReq128(Cio, X##kio.v128); \
    X##ku = LOAD64(state[14]); \
    XOReq64(Cu, X##ku); \
    X##mae.v128 = LOAD128u(state[15]); \
    X##ma = X##mae.v128; \
    X##me = GET64HI(X##mae.v128, X##mae.v128); \
    X##kame.v128 = GET64LO(X##ka, X##me); \
    XOReq128(Cae, X##mae.v128); \
    X##mio.v128 = LOAD128u(state[17]); \
    X##mi = X##mio.v128; \
    X##kemi.v128 = GET64LO(X##ke, X##mi); \
    X##mo = GET64HI(X##mio.v128, X##mio.v128); \
    XOReq128(Cio, X##mio.v128); \
    X##mu = LOAD64(state[19]); \
    XOReq64(Cu, X##mu); \
    X##sae.v128 = LOAD128(state[20]); \
    X##sa = X##sae.v128; \
    X##se = GET64HI(X##sae.v128, X##sae.v128); \
    XOReq128(Cae, X##sae.v128); \
    X##sio.v128 = LOAD128(state[22]); \
    X##si = X##sio.v128; \
    X##so = GET64HI(X##sio.v128, X##sio.v128); \
    XOReq128(Cio, X##sio.v128); \
    X##su = LOAD64(state[24]); \
    XOReq64(Cu, X##su); \

#define copyToState(state, X) \
    state[ 0] = A##bage.v64[0]; \
    state[ 1] = A##begi.v64[0]; \
    STORE64(state[ 2], X##bi); \
    STORE64(state[ 3], X##bo); \
    STORE64(state[ 4], X##bu); \
    STORE64(state[ 5], X##ga); \
    state[ 6] = A##bage.v64[1]; \
    state[ 7] = A##begi.v64[1]; \
    STORE64(state[ 8], X##go); \
    STORE64(state[ 9], X##gu); \
    state[10] = X##kame.v64[0]; \
    state[11] = X##kemi.v64[0]; \
    STORE64(state[12], X##ki); \
    STORE64(state[13], X##ko); \
    STORE64(state[14], X##ku); \
    STORE64(state[15], X##ma); \
    state[16] = X##kame.v64[1]; \
    state[17] = X##kemi.v64[1]; \
    STORE64(state[18], X##mo); \
    STORE64(state[19], X##mu); \
    STORE64(state[20], X##sa); \
    STORE64(state[21], X##se); \
    STORE64(state[22], X##si); \
    STORE64(state[23], X##so); \
    STORE64(state[24], X##su); \

#define copyStateVariables(X, Y) \
    X##bage = Y##bage; \
    X##begi = Y##begi; \
    X##bi = Y##bi; \
    X##bo = Y##bo; \
    X##bu = Y##bu; \
    X##ga = Y##ga; \
    X##go = Y##go; \
    X##gu = Y##gu; \
    X##kame = Y##kame; \
    X##kemi = Y##kemi; \
    X##ki = Y##ki; \
    X##ko = Y##ko; \
    X##ku = Y##ku; \
    X##ma = Y##ma; \
    X##mo = Y##mo; \
    X##mu = Y##mu; \
    X##sa = Y##sa; \
    X##se = Y##se; \
    X##si = Y##si; \
    X##so = Y##so; \
    X##su = Y##su; \

