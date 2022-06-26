package main

import (
	"fmt"
)


type TLBEntry struct {
	page_start uint16
	page_frame_start uint16
}

type TLB struct {
	entries [32]TLBEntry
	head int
}

// page table entry
type PTE struct {
	present bool
	dirty bool
	accessed bool
	page_frame_start uint16
	address_in_swap uint16
}

// page directory entry
type PDE struct {
	present bool
	dirty bool
	accessed bool
	page_table [8]PTE
}

type MMU struct {
	PAGE_SIZE uint16
	cr3 [8]PDE // cpu control register 3
	tlb TLB
	memory       []float64
	page_index_head uint16
	swap           []float64
}

func create_mmu_instance() MMU {
	var page_directory [8]PDE
	for i := 0; i<8; i++ {
		var page_table [8]PTE
		for x := 0; x<8; x++ {
			var address_in_swap uint16 = uint16(i*x*1024)
			pte := PTE{present: false, dirty: false, accessed: false, address_in_swap: address_in_swap, }
			page_table[x] = pte
		}
		page_directory[i] = PDE{present: false, dirty: false, accessed: false, page_table: page_table}
	}
	tlb := TLB{head: 31}
	mmu := MMU{PAGE_SIZE: 1024, cr3: page_directory, tlb: tlb, memory: make([]float64, 4096), swap: make([]float64, 65536)}
	mmu.cr3[0].page_table[0].present = true
	mmu.cr3[0].page_table[0].page_frame_start = 0

	mmu.cr3[0].page_table[1].present = true
	mmu.cr3[0].page_table[1].page_frame_start = 1024

	mmu.cr3[0].page_table[1].present = true
	mmu.cr3[0].page_table[1].page_frame_start = 2048

	mmu.cr3[0].page_table[1].present = true
	mmu.cr3[0].page_table[1].page_frame_start = 3072

	return mmu
}

func (tlb *TLB) update_tlb(page_start uint16, page_frame_start uint16) {
	// We are using a simple FIFO algorithm for TLB replacement policy
	tlb.entries[tlb.head] = TLBEntry{page_start: page_start, page_frame_start: page_frame_start}
	tlb.head = (tlb.head+1) % len(tlb.entries)
}

func (mmu *MMU) select_page_to_replace() *PTE {
	// We are using the Second-chance (FIFO with reference bit) algorithm for page replacement
	for {
		pd_index := (mmu.page_index_head>>3) & 7
		pt_index := mmu.page_index_head & 7
		pte := mmu.cr3[pd_index].page_table[pt_index]

		if pte.present && !pte.accessed {
			return &pte
		}

		fmt.Println(pd_index, pt_index)

		pte.accessed = false
		mmu.page_index_head = (mmu.page_index_head+1)%64
	}
}


func (mmu *MMU) copy_page_to_swap(pte PTE){
	var i uint16 = 0
	for i = 0; i < mmu.PAGE_SIZE; i++ {
		mmu.swap[pte.address_in_swap+i] = mmu.memory[pte.page_frame_start+i]
	}
}

func (mmu *MMU) copy_page_from_swap(pte_in_swap PTE, pte_in_pm PTE){
	var i uint16 = 0
	for i = 0; i < mmu.PAGE_SIZE; i++ {
		mmu.memory[pte_in_pm.page_frame_start+i] = mmu.swap[pte_in_swap.address_in_swap+i]
	}
}

func (mmu *MMU) vm_to_pm(vm_address uint16) uint16 {
	vm_page_start := ^(mmu.PAGE_SIZE-1) & vm_address
	vm_page_offset := (mmu.PAGE_SIZE-1) & vm_address

	// 1. Check tlb
	for _, tlb_entry := range mmu.tlb.entries {
		if tlb_entry.page_start == vm_page_start {
			// Matching entry found
			return tlb_entry.page_frame_start + vm_page_offset
		}
	}

	// TLB MISS!
	// we have to simulate cpu page table walk starting from cr3
	pd_index := (vm_address>>13) & 7
	pt_index := (vm_address>>10) & 7
	pte := mmu.cr3[pd_index].page_table[pt_index]
	if pte.present {
		// page found. Update flags, update tlb and return
		mmu.cr3[pd_index].accessed = true
		mmu.cr3[pd_index].page_table[pt_index].accessed = true
		mmu.tlb.update_tlb(vm_page_start, pte.page_frame_start)
		return pte.page_frame_start + vm_page_offset
	}

	// PAGE FAULT! bring page frame to pm and update tlb
	fmt.Println("PAGE FAULT! vm_page_start: " + fmt.Sprintf("%d", vm_page_start))
	pte_to_swap := mmu.select_page_to_replace()
	if pte_to_swap.dirty {
		mmu.copy_page_to_swap(pte)
		pte.dirty = false
	}
	mmu.copy_page_from_swap(pte, *pte_to_swap)
	pte_to_swap.present = false
	pte.present = true
	pte.page_frame_start = pte_to_swap.page_frame_start
	mmu.tlb.update_tlb(vm_page_start, pte.page_frame_start)

	return mmu.vm_to_pm(vm_address)
}



func (mmu *MMU) read(index int) float64 {
	return mmu.memory[mmu.vm_to_pm(uint16(index))]
}

func (mmu *MMU) write(index int, value float64){
	fmt.Println("Write (" + fmt.Sprintf("%d", index) + ") = " + fmt.Sprintf("%f", value))
	mmu.memory[mmu.vm_to_pm(uint16(index))] = value
}