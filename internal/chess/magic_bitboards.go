package chess

import "math/bits"

var rookMagicNumbers = [64]uint64{
	0x0080001820804000, 0x0040004020001004, 0x0200208042001008, 0x0100081000050021,
	0x0200042010020008, 0x120002002c181045, 0x0400280120921004, 0x00800d0003402080,
	0x0020801080204002, 0x1001002040011081, 0x8213004020001304, 0x0082000a00421022,
	0xc145000500080010, 0x0e96001084020008, 0x0001010001040200, 0x2060802840800900,
	0x0c2081800020400c, 0x3000404000201000, 0x8002020020804012, 0x0410018008001080,
	0x0288808008000401, 0x0001010004000802, 0x0086040010018802, 0x0a00020000a40ac5,
	0x0080004440002000, 0x8800200040005008, 0x21a0080040401000, 0x0000090100100020,
	0x0000080100110004, 0x0001000300080400, 0x2080500400186b22, 0x1020304200008401,
	0x0200400024800089, 0x4020402002401009, 0x0000802004801004, 0x0030008008080100,
	0x1002002012000904, 0x0090800400800200, 0x0014021004000108, 0x8000244882000104,
	0x0002824000228000, 0x0100201000404001, 0x0181004020010010, 0x0060100061030008,
	0x0000080011010004, 0x1080020004008080, 0x1000010208040010, 0x2040041142860021,
	0x0880004000200040, 0x140300e8820c4600, 0x1081007e40200100, 0x0100801000080080,
	0x1408008008040080, 0x2240040080020080, 0x2800880142100400, 0x160210508d040200,
	0x8000120484204102, 0xd422030082281042, 0x0005108009204202, 0x2008081000042101,
	0x0002000850218402, 0x0011000802040001, 0x2108009058020104, 0x0001010406482282,
}

var bishopMagicNumbers = [64]uint64{
	0x5404208200410100, 0x0020020208411900, 0x452109040a801000, 0x0004240180008103,
	0x0404050442008089, 0x0428441004014804, 0x01030c0212420080, 0x0001804050104811,
	0x0000420202240500, 0x2000a02210860090, 0x0800440434004010, 0x2008024081000238,
	0x0021640420880020, 0x0001220250046120, 0x8880058804132021, 0x4420008618010c90,
	0x02200006204c2110, 0x0011020810208492, 0x0008000108010016, 0x02e4001844006a00,
	0x8040808400e00008, 0x40c4101202020108, 0x0236003108310400, 0x1400480424122800,
	0x2d88880820200128, 0x0802020810100210, 0x0042240020810400, 0x0040040002110010,
	0x8200840030802000, 0x4090010241808880, 0x020803a021048843, 0x00021200008c4502,
	0x0084210811241082, 0x00208a3120600400, 0x0406482200100408, 0x0100020080980080,
	0x1910120080c49004, 0x0820008900148044, 0x0021010c08011404, 0x20608a0480284c14,
	0x008088143080401f, 0x00222090041a0840, 0x0002424020801000, 0x8000002019014800,
	0x0400040094000200, 0x200802004a001410, 0x00200400821a00a0, 0x0402080052801300,
	0x0006091120100204, 0x0000220802080000, 0x0178804404048003, 0x0040084084040000,
	0x0820c01020220858, 0x080c090828082422, 0x0020a00421304004, 0x04103410c4004402,
	0x28a1c04404201208, 0x0080042404040480, 0x0000019954040430, 0x000c81a000840420,
	0x0008020822c55400, 0x0808280820188082, 0x0c02092004040042, 0x0040140404022821,
}

var (
	rookMagicMasks     [64]uint64
	bishopMagicMasks   [64]uint64
	rookMagicShifts    [64]uint8
	bishopMagicShifts  [64]uint8
	rookMagicAttacks   [64][]uint64
	bishopMagicAttacks [64][]uint64
)

func initMagicBitboards() {
	for sq := int8(0); sq < 64; sq++ {
		rookMask := magicRelevantMask(sq, Rook)
		bishopMask := magicRelevantMask(sq, Bishop)
		rookMagicMasks[sq] = rookMask
		bishopMagicMasks[sq] = bishopMask
		rookMagicShifts[sq] = uint8(64 - bits.OnesCount64(rookMask))
		bishopMagicShifts[sq] = uint8(64 - bits.OnesCount64(bishopMask))
		rookMagicAttacks[sq] = buildMagicAttackTable(sq, rookMask, rookMagicNumbers[sq], rookMagicShifts[sq], Rook)
		bishopMagicAttacks[sq] = buildMagicAttackTable(sq, bishopMask, bishopMagicNumbers[sq], bishopMagicShifts[sq], Bishop)
	}
}

func magicRelevantMask(square int8, pieceType int8) uint64 {
	mask := uint64(0)
	startDir, endDir := 0, 8
	if pieceType == Bishop {
		startDir, endDir = 4, 8
	} else if pieceType == Rook {
		startDir, endDir = 0, 4
	}

	for dirIdx := startDir; dirIdx < endDir; dirIdx++ {
		ray := sliderAttackMasks[square][dirIdx]
		for ray != 0 {
			nextBit := ray & -ray
			nextSquare := int8(bits.TrailingZeros64(nextBit))
			ray &^= nextBit
			if sliderAttackMasks[nextSquare][dirIdx] == 0 {
				continue
			}
			mask |= nextBit
		}
	}

	return mask
}

func magicOccupancyFromIndex(index int, mask uint64) uint64 {
	occ := uint64(0)
	bitPos := 0
	for mask != 0 {
		bit := mask & -mask
		mask &^= bit
		if index&(1<<bitPos) != 0 {
			occ |= bit
		}
		bitPos++
	}
	return occ
}

func buildMagicAttackTable(square int8, mask, magic uint64, shift uint8, pieceType int8) []uint64 {
	size := 1 << bits.OnesCount64(mask)
	table := make([]uint64, size)
	for i := 0; i < size; i++ {
		occ := magicOccupancyFromIndex(i, mask)
		index := int(((occ & mask) * magic) >> shift)
		table[index] = sliderAttacksFromOcc(square, occ, pieceType)
	}
	return table
}

func sliderAttacksFromOcc(square int8, occ uint64, pieceType int8) uint64 {
	var attacks uint64
	startDir, endDir := 0, 8
	if pieceType == Bishop {
		startDir, endDir = 4, 8
	} else if pieceType == Rook {
		startDir, endDir = 0, 4
	}

	for dirIdx := startDir; dirIdx < endDir; dirIdx++ {
		ray := sliderAttackMasks[square][dirIdx]
		if ray == 0 {
			continue
		}
		blocker := firstBlockerOnRay(occ, ray, rayDirections[dirIdx])
		if blocker == 0 {
			attacks |= ray
			continue
		}
		blockerIdx := int8(bits.TrailingZeros64(blocker))
		attacks |= ray ^ sliderAttackMasks[blockerIdx][dirIdx]
	}

	return attacks
}

func rookAttacksMagic(square int8, occ uint64) uint64 {
	mask := rookMagicMasks[square]
	return rookMagicAttacks[square][((occ&mask)*rookMagicNumbers[square])>>rookMagicShifts[square]]
}

func bishopAttacksMagic(square int8, occ uint64) uint64 {
	mask := bishopMagicMasks[square]
	return bishopMagicAttacks[square][((occ&mask)*bishopMagicNumbers[square])>>bishopMagicShifts[square]]
}
